package bili

import (
	"fmt"
	"os"
	"regexp"
)

var ugcSeasonTitlePattern = regexp.MustCompile(`(?s)"ugc_season"\s*:\s*\{.*?"title"\s*:\s*"([^"]*)"`)

// GetUGCSeasonTitle tries to read ugc_season.title from a saved initial-state file.
// It prefers parsed JSON and falls back to raw-text extraction when the file is not strict JSON.
func (b *Bili) GetUGCSeasonTitle() (string, error) {
	jsInfo, _, err := b.ParseJSON(InitialStatePath)
	if err == nil {
		if title, ok := b.GetNestedString(jsInfo, "ugc_season", "title"); ok {
			return title, nil
		}
	}

	raw, readErr := os.ReadFile(InitialStatePath)
	if readErr != nil {
		if err != nil {
			return "", fmt.Errorf("parse failed: %v; read failed: %v", err, readErr)
		}
		return "", readErr
	}

	matches := ugcSeasonTitlePattern.FindStringSubmatch(string(raw))
	if len(matches) < 2 {
		if err != nil {
			return "", fmt.Errorf("ugc_season.title not found; parse failed: %v", err)
		}
		return "", fmt.Errorf("ugc_season.title not found")
	}

	return matches[1], nil
}
