package search

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
)

type CaseSensitivity int

const (
	CaseSmart CaseSensitivity = iota
	CaseSensitive
	CaseInsensitive
)

type Match struct {
	Path       string
	LineNumber int
	LineText   string
	Submatches []Submatch
}

type Submatch struct {
	Match string
	Start int
	End   int
}

type RipgrepMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MatchData struct {
	Path struct {
		Text string `json:"text"`
	} `json:"path"`
	Lines struct {
		Text string `json:"text"`
	} `json:"lines"`
	LineNumber int `json:"line_number"`
	Submatches []struct {
		Match struct {
			Text string `json:"text"`
		} `json:"match"`
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"submatches"`
}

type Searcher struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func NewSearcher() *Searcher {
	return &Searcher{}
}

func (s *Searcher) Search(ctx context.Context, pattern, path string, caseSensitivity CaseSensitivity, results chan<- Match) error {
	if pattern == "" {
		close(results)
		return nil
	}

	args := []string{
		"--json",
		"--line-number",
		"--column",
		"--max-count=1000",
		"--",
		pattern,
	}

	// Add case sensitivity flag based on mode
	switch caseSensitivity {
	case CaseSmart:
		args = append(args[:4], append([]string{"--smart-case"}, args[4:]...)...)
	case CaseSensitive:
		args = append(args[:4], append([]string{"--case-sensitive"}, args[4:]...)...)
	case CaseInsensitive:
		args = append(args[:4], append([]string{"--ignore-case"}, args[4:]...)...)
	}

	if path != "" {
		args = append(args, path)
	} else {
		args = append(args, ".")
	}

	s.cmd = exec.CommandContext(ctx, "rg", args...)

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		close(results)
		return err
	}

	if err := s.cmd.Start(); err != nil {
		close(results)
		return err
	}

	go func() {
		defer close(results)
		scanner := bufio.NewScanner(stdout)

		// Buffer size 1MB for long lines (ripgrep can return very long matches)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			var msg RipgrepMessage
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				continue
			}

			if msg.Type != "match" {
				continue
			}

			var matchData MatchData
			if err := json.Unmarshal(msg.Data, &matchData); err != nil {
				continue
			}

			match := Match{
				Path:       matchData.Path.Text,
				LineNumber: matchData.LineNumber,
				LineText:   matchData.Lines.Text,
			}

			for _, sm := range matchData.Submatches {
				match.Submatches = append(match.Submatches, Submatch{
					Match: sm.Match.Text,
					Start: sm.Start,
					End:   sm.End,
				})
			}

			select {
			case results <- match:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		s.cmd.Wait()
	}()

	return nil
}

func (s *Searcher) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}
}

type FileContext struct {
	Lines      []string
	StartLine  int
	MatchLine  int
	MatchStart int
	MatchEnd   int
	Submatches []Submatch
}

func GetFileContextWithMatches(path string, lineNum, contextLines int, submatches []Submatch) (*FileContext, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	startLine := lineNum - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := lineNum + contextLines

	scanner := bufio.NewScanner(file)
	var lines []string
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		if currentLine < startLine {
			continue
		}
		if currentLine > endLine {
			break
		}
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &FileContext{
		Lines:      lines,
		StartLine:  startLine,
		MatchLine:  lineNum,
		Submatches: submatches,
	}, nil
}

func GetFileContext(path string, lineNum, contextLines int) (*FileContext, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	startLine := lineNum - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := lineNum + contextLines

	scanner := bufio.NewScanner(file)
	var lines []string
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		if currentLine < startLine {
			continue
		}
		if currentLine > endLine {
			break
		}
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &FileContext{
		Lines:     lines,
		StartLine: startLine,
		MatchLine: lineNum,
	}, nil
}
