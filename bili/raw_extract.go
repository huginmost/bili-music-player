package bili

import (
	"fmt"
	"strconv"
	"strings"
)

func findKeyValueStart(raw, key string) (int, error) {
	pattern := `"` + key + `":`
	idx := strings.Index(raw, pattern)
	if idx < 0 {
		return -1, fmt.Errorf("%s not found", key)
	}

	return idx + len(pattern), nil
}

func findMatchingBracket(raw string, start int, openCh, closeCh byte) (int, error) {
	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(raw); i++ {
		ch := raw[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			continue
		}

		if ch == openCh {
			depth++
			continue
		}

		if ch == closeCh {
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}

	return -1, fmt.Errorf("matching bracket not found")
}

func extractObjectForKey(raw, key string) (string, error) {
	start, err := findKeyValueStart(raw, key)
	if err != nil {
		return "", err
	}

	for start < len(raw) && (raw[start] == ' ' || raw[start] == '\n' || raw[start] == '\r' || raw[start] == '\t') {
		start++
	}
	if start >= len(raw) || raw[start] != '{' {
		return "", fmt.Errorf("%s is not an object", key)
	}

	end, err := findMatchingBracket(raw, start, '{', '}')
	if err != nil {
		return "", err
	}

	return raw[start : end+1], nil
}

func extractArrayForKey(raw, key string) (string, error) {
	start, err := findKeyValueStart(raw, key)
	if err != nil {
		return "", err
	}

	for start < len(raw) && (raw[start] == ' ' || raw[start] == '\n' || raw[start] == '\r' || raw[start] == '\t') {
		start++
	}
	if start >= len(raw) || raw[start] != '[' {
		return "", fmt.Errorf("%s is not an array", key)
	}

	end, err := findMatchingBracket(raw, start, '[', ']')
	if err != nil {
		return "", err
	}

	return raw[start : end+1], nil
}

func extractStringForKey(raw, key string) (string, error) {
	start, err := findKeyValueStart(raw, key)
	if err != nil {
		return "", err
	}

	for start < len(raw) && (raw[start] == ' ' || raw[start] == '\n' || raw[start] == '\r' || raw[start] == '\t') {
		start++
	}
	if start >= len(raw) || raw[start] != '"' {
		return "", fmt.Errorf("%s is not a string", key)
	}

	inString := true
	escaped := false
	for i := start + 1; i < len(raw); i++ {
		ch := raw[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if inString && ch == '"' {
			value, err := strconv.Unquote(raw[start : i+1])
			if err == nil {
				return value, nil
			}
			return raw[start+1 : i], nil
		}
	}

	return "", fmt.Errorf("%s string not terminated", key)
}

func splitTopLevelObjects(rawArray string) ([]string, error) {
	rawArray = strings.TrimSpace(rawArray)
	if len(rawArray) < 2 || rawArray[0] != '[' || rawArray[len(rawArray)-1] != ']' {
		return nil, fmt.Errorf("not an array")
	}

	var items []string
	body := rawArray[1 : len(rawArray)-1]
	inString := false
	escaped := false
	depth := 0
	start := -1

	for i := 0; i < len(body); i++ {
		ch := body[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			continue
		}

		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
			continue
		}

		if ch == '}' {
			depth--
			if depth == 0 && start >= 0 {
				items = append(items, body[start:i+1])
				start = -1
			}
		}
	}

	return items, nil
}
