package bili

import (
	"encoding/json"
	"fmt"
	"os"
)

type BMPInfoItem struct {
	Title string `json:"title"`
	Pic   string `json:"pic"`
	BVID  string `json:"bvid"`
	Audio string `json:"audio"`
}

type BMPInfoPayload struct {
	PLInfo []BMPInfoItem `json:"plinfo"`
}

// ReadBMPInfo reads bmpinfo.json and returns the exported playlist payload.
func (b *Bili) ReadBMPInfo() (BMPInfoPayload, error) {
	raw, err := os.ReadFile(BMPInfoPath)
	if err != nil {
		return BMPInfoPayload{}, err
	}

	var payload BMPInfoPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return BMPInfoPayload{}, err
	}

	return payload, nil
}

func (b *Bili) writeBMPInfo(payload BMPInfoPayload) error {
	formatted, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(BMPInfoPath, formatted, 0o644)
}

// GetBMPInfo reads the saved initial-state data, extracts the first section's episodes,
// and writes a simplified playlist payload to bmpinfo.json.
func (b *Bili) GetBMPInfo() ([]BMPInfoItem, error) {
	raw, err := os.ReadFile(InitialStatePath)
	if err != nil {
		return nil, err
	}

	ugcSeason, err := extractObjectForKey(string(raw), "ugc_season")
	if err != nil {
		return nil, err
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

	episodeObjects, err := splitTopLevelObjects(episodes)
	if err != nil {
		return nil, err
	}

	items := make([]BMPInfoItem, 0, len(episodeObjects))
	for _, obj := range episodeObjects {
		title, err := extractStringForKey(obj, "title")
		if err != nil {
			continue
		}
		pic, err := extractStringForKey(obj, "arc_pic")
		if err != nil {
			pic, err = extractStringForKey(obj, "pic")
			if err != nil {
				pic = ""
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

	payload := BMPInfoPayload{PLInfo: items}
	if err := b.writeBMPInfo(payload); err != nil {
		return nil, err
	}

	return items, nil
}
