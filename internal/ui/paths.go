package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	pathMaxDepth          = 5
	pathCacheTTL          = 30 * time.Second
	pathDropdownMaxHeight = 8
	maxPathResults        = 50
)

type PathEntry struct {
	Path  string
	IsDir bool
	Score int
}

type PathProvider struct {
	root      string
	cache     []PathEntry
	cacheMu   sync.RWMutex
	cacheTime time.Time
	ttl       time.Duration
}

func NewPathProvider(root string) *PathProvider {
	return &PathProvider{
		root: root,
		ttl:  pathCacheTTL,
	}
}

func (p *PathProvider) LoadPaths() []PathEntry {
	p.cacheMu.RLock()
	if time.Since(p.cacheTime) < p.ttl && p.cache != nil {
		defer p.cacheMu.RUnlock()
		return p.cache
	}
	p.cacheMu.RUnlock()

	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()

	// Re-check after acquiring lock
	if time.Since(p.cacheTime) < p.ttl && p.cache != nil {
		return p.cache
	}

	paths := []PathEntry{}
	p.walkDirectory(p.root, 0, &paths)

	p.cache = paths
	p.cacheTime = time.Now()
	return paths
}

func (p *PathProvider) walkDirectory(root string, depth int, paths *[]PathEntry) {
	if depth >= pathMaxDepth {
		return
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files and common large directories
		if strings.HasPrefix(name, ".") ||
			name == "node_modules" ||
			name == "vendor" ||
			name == ".git" {
			continue
		}

		path := filepath.Join(root, name)
		if root == "." {
			path = name
		}

		isDir := entry.IsDir()
		*paths = append(*paths, PathEntry{
			Path:  path,
			IsDir: isDir,
		})

		if isDir {
			p.walkDirectory(path, depth+1, paths)
		}
	}
}

func (p *PathProvider) FilterPaths(input string, allPaths []PathEntry) []PathEntry {
	if input == "" {
		return nil
	}

	var matches []PathEntry
	for _, entry := range allPaths {
		score := scorePathMatch(input, entry.Path)
		if score > 0 {
			entry.Score = score
			matches = append(matches, entry)
		}
	}

	sortPathMatches(matches)

	if len(matches) > maxPathResults {
		matches = matches[:maxPathResults]
	}

	return matches
}

func scorePathMatch(input, path string) int {
	// Normalize to forward slashes for consistent matching across platforms
	input = strings.ToLower(filepath.ToSlash(input))
	pathLower := strings.ToLower(filepath.ToSlash(path))

	if input == pathLower {
		return 2000
	}

	if strings.HasPrefix(pathLower, input) {
		return 1000
	}

	filename := filepath.Base(pathLower)
	if strings.HasPrefix(filename, input) {
		return 800
	}

	// Boundary match (e.g., "ui" matching "internal/ui")
	if strings.Contains(pathLower, "/"+input) {
		return 500
	}

	if strings.Contains(pathLower, input) {
		return 100
	}

	return 0
}

func sortPathMatches(matches []PathEntry) {
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].Path < matches[j].Path
	})
}
