package bili

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"

var playInfoPattern = regexp.MustCompile(`(?s)window\.__playinfo__\s*=\s*(\{.*?\})\s*</script>`)
var initialStatePattern = regexp.MustCompile(`(?s)window\.__INITIAL_STATE__\s*=\s*(\{.*?\})\s*;\s*\(function`)

func (b *Bili) fetchVideoPage(bvid string) (string, error) {
	if b == nil {
		return "", fmt.Errorf("bili is nil")
	}

	videoURL := fmt.Sprintf("https://www.bilibili.com/video/%s", bvid)
	req, err := http.NewRequest(http.MethodGet, videoURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Referer", baseURL)

	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func extractScriptJSON(body string, pattern *regexp.Regexp, name string) (string, error) {
	matches := pattern.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("%s not found in response", name)
	}

	return strings.TrimSpace(matches[1]), nil
}

func writeJSONFile(outputPath, content string) error {
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(outputPath, []byte(content), 0o644)
}

// bili_get_pi fetches a Bilibili video page, extracts window.__playinfo__, and writes it to file.
func (b *Bili) bili_get_pi(bvid, outputPath string) (string, error) {
	body, err := b.fetchVideoPage(bvid)
	if err != nil {
		return "", err
	}

	playInfo, err := extractScriptJSON(body, playInfoPattern, "window.__playinfo__")
	if err != nil {
		return "", err
	}

	if err := writeJSONFile(outputPath, playInfo); err != nil {
		return "", err
	}

	return playInfo, nil
}

// bili_get_is fetches a Bilibili video page, extracts window.__INITIAL_STATE__, and writes it to file.
func (b *Bili) bili_get_is(bvid, outputPath string) (string, error) {
	body, err := b.fetchVideoPage(bvid)
	if err != nil {
		return "", err
	}

	initialState, err := extractScriptJSON(body, initialStatePattern, "window.__INITIAL_STATE__")
	if err != nil {
		return "", err
	}

	if err := writeJSONFile(outputPath, initialState); err != nil {
		return "", err
	}

	return initialState, nil
}

// bili_js reads a saved playinfo JSON file, formats it, and returns the parsed data.
func (b *Bili) bili_js(inputPath string) (map[string]any, string, error) {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, "", err
	}

	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, "", err
	}

	formatted, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return nil, "", err
	}

	return parsed, string(formatted), nil
}
