package bili

import (
	"encoding/json"
	"os"
	"strings"
)

type PlayerSettings struct {
	ActivePlaylistTitle string  `json:"activePlaylistTitle"`
	ActiveTrackBVID     string  `json:"activeTrackBvid"`
	CurrentTime         float64 `json:"currentTime"`
	PlayMode            string  `json:"playMode"`
	Volume              float64 `json:"volume"`
	ShuffleQueue        []int   `json:"shuffleQueue"`
	HistoryStack        []int   `json:"historyStack"`
}

func defaultSettings() PlayerSettings {
	return PlayerSettings{
		PlayMode:     "sequence",
		Volume:       0.72,
		ShuffleQueue: []int{},
		HistoryStack: []int{},
	}
}

func (b *Bili) ReadSettings() (PlayerSettings, error) {
	raw, err := os.ReadFile(SettingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultSettings(), nil
		}
		return PlayerSettings{}, err
	}

	if strings.TrimSpace(string(raw)) == "" {
		return defaultSettings(), nil
	}

	settings := defaultSettings()
	if err := json.Unmarshal(raw, &settings); err != nil {
		return PlayerSettings{}, err
	}

	if settings.PlayMode == "" {
		settings.PlayMode = "sequence"
	}
	if settings.CurrentTime < 0 {
		settings.CurrentTime = 0
	}
	if settings.Volume < 0 || settings.Volume > 1 {
		settings.Volume = 0.72
	}
	if settings.ShuffleQueue == nil {
		settings.ShuffleQueue = []int{}
	}
	if settings.HistoryStack == nil {
		settings.HistoryStack = []int{}
	}

	return settings, nil
}

func (b *Bili) WriteSettings(settings PlayerSettings) error {
	if settings.PlayMode == "" {
		settings.PlayMode = "sequence"
	}
	if settings.CurrentTime < 0 {
		settings.CurrentTime = 0
	}
	if settings.Volume < 0 {
		settings.Volume = 0
	}
	if settings.Volume > 1 {
		settings.Volume = 1
	}
	if settings.ShuffleQueue == nil {
		settings.ShuffleQueue = []int{}
	}
	if settings.HistoryStack == nil {
		settings.HistoryStack = []int{}
	}

	formatted, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(SettingsPath, formatted, 0o644)
}
