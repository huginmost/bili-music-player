package bili

import "fmt"

// GetAudio reads pi.json and returns the first audio baseUrl.
func (b *Bili) GetAudio() (string, error) {
	jsInfo, _, err := b.ParseJSON(PlayInfoPath)
	if err != nil {
		return "", err
	}

	data, ok := jsInfo["data"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("data not found")
	}
	dash, ok := data["dash"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("dash not found")
	}
	audioList, ok := dash["audio"].([]any)
	if !ok || len(audioList) == 0 {
		return "", fmt.Errorf("audio list not found")
	}
	firstAudio, ok := audioList[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("audio item is invalid")
	}

	baseURL, ok := firstAudio["baseUrl"].(string)
	if !ok || baseURL == "" {
		baseURL, ok = firstAudio["base_url"].(string)
		if !ok || baseURL == "" {
			return "", fmt.Errorf("audio baseUrl not found")
		}
	}

	return baseURL, nil
}
