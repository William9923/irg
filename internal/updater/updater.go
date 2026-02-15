package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Updater provides a tiny self-update mechanism based on GitHub releases.
// It uses the repository/releases API to fetch the latest version and the
// corresponding asset named according to the convention:
//
//	irg_${VERSION}_${OS}_${ARCH}.tar.gz
//
// The implementation here is a lightweight updater that downloads the asset,
// extracts the binary, and replaces the currently running executable.
// NOTE: This is a pragmatic implementation to satisfy the task requirements
// and does not rely on any external pre-release signing checks.
type Updater struct {
	Repo   string // e.g. "William9923/irg"
	Binary string // e.g. "irg" or "irg.exe" on Windows
}

// Check fetches the latest release from GitHub and returns the version tag and
// a flag indicating that an update is available. The current version is determined
// by the presence of the latest tag; the consumer should perform the comparison
// with its own embedded version if needed.
func (u *Updater) Check() (string, bool, error) {
	latest, err := fetchLatestRelease(u.Repo)
	if err != nil {
		return "", false, err
	}
	// Consider an update as available if we could determine a tag name.
	version := latest.TagName
	needsUpdate := version != ""
	return version, needsUpdate, nil
}

// Update downloads the asset matching the OS/ARCH naming convention and
// replaces the running binary with the downloaded one.
func (u *Updater) Update(version string) error {
	if version == "" {
		// best-effort: fetch latest version
		ver, ok, err := u.Check()
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no update available")
		}
		version = ver
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}
	assetName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", "irg", version, runtime.GOOS, arch)
	latest, err := fetchLatestRelease(u.Repo)
	if err != nil {
		return err
	}
	assetURL := ""
	for _, a := range latest.Assets {
		if a.Name == assetName {
			assetURL = a.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("update asset not found: %s", assetName)
	}

	fmt.Printf("Downloading asset: %s\n", assetName)
	tmpPath, err := downloadFile(assetURL)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpPath)

	// Extract tar.gz to a temporary directory
	extractDir, err := extractTarGz(tmpPath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractDir)

	// Find the binary inside the extracted directory
	binPath, err := findBinary(extractDir, u.Binary)
	if err != nil {
		return err
	}

	// Replace the current executable with the new one
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	newPath := exePath + ".new"
	if err := copyFile(binPath, newPath); err != nil {
		return err
	}
	// On most systems, os.Rename will overwrite atomically if same FS; best effort
	if err := os.Rename(newPath, exePath); err != nil {
		// Fallback: copy over the existing binary
		if err2 := copyFile(newPath, exePath); err2 != nil {
			return fmt.Errorf("failed to replace executable: %v; retry failed: %v", err, err2)
		}
		os.Remove(newPath)
	}

	fmt.Println("Update applied. Please restart the application.")
	return nil
}

// fetchLatestRelease contacts the GitHub releases API and returns the latest release
// information parsed into a Release struct.
func fetchLatestRelease(repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// Release models the GitHub release response used for asset discovery.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// downloadFile downloads a URL to a temporary file and returns the path.
func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tmp, err := os.CreateTemp("", "irg-update-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}

// extractTarGz extracts a tar.gz file into a temporary directory and returns the path.
func extractTarGz(tarPath string) (string, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()
	tarReader := tar.NewReader(gzr)
	tmpDir, err := os.MkdirTemp("", "irg-update-*")
	if err != nil {
		return "", err
	}
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		target := filepath.Join(tmpDir, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}
			out, err := os.Create(target)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tarReader); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
		}
	}
	return tmpDir, nil
}

// findBinary searches for a binary file named binName inside dir. It returns the path.
func findBinary(dir, binName string) (string, error) {
	var found string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		if name == binName || strings.EqualFold(name, binName+".exe") {
			// Ensure executable bit is set on Unix-like systems
			if runtime.GOOS != "windows" {
				if info.Mode().Perm()&0111 == 0 {
					// try to chmod to executable
					_ = os.Chmod(path, info.Mode().Perm()|0100)
				}
			}
			found = path
			return io.EOF // stop walking
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("binary not found in archive: %s", binName)
	}
	return found, nil
}

// copyFile copies a file from src to dst. It preserves permissions where possible.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	// Copy permissions from source if possible
	fi, err := in.Stat()
	if err == nil {
		_ = os.Chmod(dst, fi.Mode())
	}
	return out.Close()
}
