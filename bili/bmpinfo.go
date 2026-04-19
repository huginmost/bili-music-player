package bili

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type BMPInfoItem struct {
	Title string `json:"title"`
	Pic   string `json:"pic"`
	BVID  string `json:"bvid"`
	Audio string `json:"audio"`
}

type BMPInfoPayload map[string][]BMPInfoItem

// ReadBMPInfo reads bmpinfo.json and returns the exported playlist payload.
func (b *Bili) ReadBMPInfo() (BMPInfoPayload, error) {
	raw, err := os.ReadFile(BMPInfoPath)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(string(raw)) == "" {
		return BMPInfoPayload{}, nil
	}

	var payload BMPInfoPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	if payload == nil {
		payload = BMPInfoPayload{}
	}

	return payload, nil
}

func (b *Bili) writeBMPInfo(payload BMPInfoPayload) error {
	if payload == nil {
		payload = BMPInfoPayload{}
	}

	formatted, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(BMPInfoPath, formatted, 0o644)
}

func (b *Bili) readInitialStateString(key string) (string, error) {
	jsInfo, _, err := b.ParseJSON(InitialStatePath)
	if err == nil {
		if value, ok := b.GetNestedString(jsInfo, key); ok {
			return value, nil
		}
	}

	raw, readErr := os.ReadFile(InitialStatePath)
	if readErr != nil {
		if err != nil {
			return "", fmt.Errorf("parse failed: %v; read failed: %v", err, readErr)
		}
		return "", readErr
	}

	value, extractErr := extractStringForKey(string(raw), key)
	if extractErr != nil {
		if err != nil {
			return "", fmt.Errorf("%s not found; parse failed: %v", key, err)
		}
		return "", extractErr
	}

	return value, nil
}

func normalizeBMPInfoPic(pic string) string {
	if strings.HasPrefix(pic, "//") {
		return "http:" + pic
	}

	return pic
}

func extractBMPInfoItemsFromArray(arrayRaw string, titleKey string, picKeys ...string) ([]BMPInfoItem, error) {
	itemObjects, err := splitTopLevelObjects(arrayRaw)
	if err != nil {
		return nil, err
	}

	items := make([]BMPInfoItem, 0, len(itemObjects))
	for _, obj := range itemObjects {
		title, err := extractStringForKey(obj, titleKey)
		if err != nil {
			continue
		}

		pic := ""
		for _, key := range picKeys {
			value, picErr := extractStringForKey(obj, key)
			if picErr == nil {
				pic = normalizeBMPInfoPic(value)
				break
			}
		}

		bvid, err := extractStringForKey(obj, "bvid")
		if err != nil {
			continue
		}

		items = append(items, BMPInfoItem{
			Title: title,
			Pic:   pic,
			BVID:  bvid,
			Audio: "",
		})
	}

	return items, nil
}

func (b *Bili) extractBMPInfoItems() ([]BMPInfoItem, error) {
	raw, err := os.ReadFile(InitialStatePath)
	if err != nil {
		return nil, err
	}

	ugcSeason, err := extractObjectForKey(string(raw), "ugc_season")
	if err != nil {
		return nil, nil
	}

	sections, err := extractArrayForKey(ugcSeason, "sections")
	if err != nil {
		return nil, err
	}

	sectionObjects, err := splitTopLevelObjects(sections)
	if err != nil {
		return nil, err
	}
	if len(sectionObjects) == 0 {
		return nil, fmt.Errorf("ugc_season.sections is empty")
	}

	episodes, err := extractArrayForKey(sectionObjects[0], "episodes")
	if err != nil {
		return nil, err
	}

	return extractBMPInfoItemsFromArray(episodes, "title", "arc_pic", "pic")
}

func (b *Bili) extractListBMPInfoItems() ([]BMPInfoItem, error) {
	raw, err := os.ReadFile(InitialStatePath)
	if err != nil {
		return nil, err
	}

	resourceList, err := extractArrayForKey(string(raw), "resourceList")
	if err != nil {
		return nil, err
	}

	return extractBMPInfoItemsFromArray(resourceList, "title", "cover", "pic", "arc_pic")
}

func (b *Bili) upsertBMPInfo(key string, items []BMPInfoItem) ([]BMPInfoItem, error) {
	payload := BMPInfoPayload{}
	if _, err := os.Stat(BMPInfoPath); err == nil {
		payload, err = b.ReadBMPInfo()
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	if _, exists := payload[key]; exists {
		return payload[key], nil
	}

	payload[key] = items
	if err := b.writeBMPInfo(payload); err != nil {
		return nil, err
	}

	return items, nil
}

func (b *Bili) readPlaylistID() (string, error) {
	jsInfo, _, err := b.ParseJSON(InitialStatePath)
	if err == nil {
		if playlist, ok := jsInfo["playlist"].(map[string]any); ok {
			if idValue, ok := playlist["id"]; ok {
				switch v := idValue.(type) {
				case string:
					if v != "" {
						return v, nil
					}
				case float64:
					return fmt.Sprintf("%.0f", v), nil
				}
			}
		}
	}

	raw, readErr := os.ReadFile(InitialStatePath)
	if readErr != nil {
		if err != nil {
			return "", fmt.Errorf("parse failed: %v; read failed: %v", err, readErr)
		}
		return "", readErr
	}

	playlist, extractErr := extractObjectForKey(string(raw), "playlist")
	if extractErr != nil {
		if err != nil {
			return "", fmt.Errorf("playlist not found; parse failed: %v", err)
		}
		return "", extractErr
	}

	literal, extractErr := extractLiteralForKey(playlist, "id")
	if extractErr != nil {
		if err != nil {
			return "", fmt.Errorf("playlist.id not found; parse failed: %v", err)
		}
		return "", extractErr
	}

	if len(literal) >= 2 && literal[0] == '"' && literal[len(literal)-1] == '"' {
		value, unquoteErr := extractStringForKey(playlist, "id")
		if unquoteErr == nil {
			return value, nil
		}
	}

	return literal, nil
}

// GetBMPInfo reads the saved initial-state data, extracts the first section's episodes,
// and writes grouped playlist data to bmpinfo.json keyed by ugc title.
func (b *Bili) GetBMPInfo() ([]BMPInfoItem, error) {
	items, err := b.extractBMPInfoItems()
	if err != nil {
		return nil, err
	}
	if items == nil {
		return nil, nil
	}

	ugcTitle, err := b.GetUGCSeasonTitle()
	if err != nil {
		return nil, err
	}

	return b.upsertBMPInfo(ugcTitle, items)
}

// GetListBMPInfo reads the saved list initial-state data, extracts resourceList,
// and writes grouped playlist data to bmpinfo.json keyed by list title.
func (b *Bili) GetListBMPInfo() ([]BMPInfoItem, error) {
	listTitle := b.GetListTitle()
	if listTitle == "" {
		return nil, fmt.Errorf("list title not found")
	}

	items, err := b.extractListBMPInfoItems()
	if err != nil {
		return nil, err
	}

	return b.upsertBMPInfo(listTitle, items)
}
