package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func extractText(v any) []string {
	var out []string
	var walk func(any)
	walk = func(value any) {
		switch x := value.(type) {
		case map[string]any:
			if role, _ := x["role"].(string); role != "" && role != "user" && role != "assistant" {
				return
			}
			for _, key := range []string{"text", "content", "message", "prompt", "response"} {
				if nested, ok := x[key]; ok {
					walk(nested)
				}
			}
		case []any:
			for _, item := range x {
				walk(item)
			}
		case string:
			if strings.TrimSpace(x) != "" {
				out = append(out, x)
			}
		case fmt.Stringer:
			out = append(out, x.String())
		}
	}
	walk(v)
	return out
}

func cleanCandidate(text string) string {
	text = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, text)
	text = strings.Join(strings.Fields(text), " ")
	if strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[") {
		return ""
	}
	return text
}

func ReadTextCandidates(path string, limit int) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return readTextCandidatesFromDir(path, limit)
	}
	if strings.HasSuffix(path, ".jsonl") {
		return readTextCandidatesFromJSONL(path, limit)
	}
	return readTextCandidatesFromFile(path, limit)
}

func readTextCandidatesFromDir(root string, limit int) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isSearchableFile(path) {
			return nil
		}
		items, readErr := ReadTextCandidates(path, remaining(limit, len(out)))
		if readErr != nil {
			return nil
		}
		out = append(out, items...)
		if limit > 0 && len(out) >= limit {
			return filepath.SkipAll
		}
		return nil
	})
	return out, err
}

func readTextCandidatesFromFile(path string, limit int) ([]string, error) {
	if !isSearchableFile(path) {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) > 512*1024 {
		data = data[:512*1024]
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" || ext == ".yaml" || ext == ".yml" {
		var parsed any
		if err := json.Unmarshal(data, &parsed); err == nil {
			return firstClean(extractText(parsed), limit), nil
		}
	}
	return firstClean(strings.Split(string(data), "\n"), limit), nil
}

func isSearchableFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jsonl", ".json", ".md", ".txt", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func firstClean(items []string, limit int) []string {
	var out []string
	for _, item := range items {
		item = cleanCandidate(item)
		if item == "" {
			continue
		}
		out = append(out, item)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func remaining(limit, current int) int {
	if limit <= 0 {
		return 0
	}
	left := limit - current
	if left < 0 {
		return 0
	}
	return left
}
