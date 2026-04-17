package bili

import "fmt"

// GetAudio reads pi.json and returns the audio baseUrl with the highest bandwidth.
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

	bestBandwidth := -1
	bestURL := ""

	for _, item := range audioList {
		audioInfo, ok := item.(map[string]any)
		if !ok {
			continue
		}

		bandwidthValue, ok := audioInfo["bandwidth"].(float64)
		if !ok {
			continue
		}

		baseURL, ok := audioInfo["baseUrl"].(string)
		if !ok || baseURL == "" {
			baseURL, ok = audioInfo["base_url"].(string)
			if !ok || baseURL == "" {
				continue
			}
		}

		bandwidth := int(bandwidthValue)
		if bandwidth > bestBandwidth {
			bestBandwidth = bandwidth
			bestURL = baseURL
		}
	}

	if bestURL == "" {
		return "", fmt.Errorf("audio baseUrl not found")
	}

	return bestURL, nil
}
