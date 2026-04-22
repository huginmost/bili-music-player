package bili

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

// FindTrack looks up one playlist item by playlist title and bvid.
func (b *Bili) FindTrack(playlistTitle, bvid string) (BMPInfoItem, bool, error) {
	payload, err := b.ReadBMPInfo()
	if err != nil {
		return BMPInfoItem{}, false, err
	}

	items, ok := payload[playlistTitle]
	if !ok {
		return BMPInfoItem{}, false, nil
	}

	for _, item := range items {
		if item.BVID == bvid {
			return item, true, nil
		}
	}

	return BMPInfoItem{}, false, nil
}

// DeletePlaylist removes one playlist entry from bmpinfo.json.
func (b *Bili) DeletePlaylist(title string) error {
	payload, err := b.ReadBMPInfo()
	if err != nil {
		return err
	}

	if _, exists := payload[title]; !exists {
		return fmt.Errorf("playlist %q not found", title)
	}

	delete(payload, title)
	return b.writeBMPInfo(payload)
}

// DeleteTrack removes one track from a playlist and drops the playlist when it becomes empty.
func (b *Bili) DeleteTrack(playlistTitle, bvid string) error {
	payload, err := b.ReadBMPInfo()
	if err != nil {
		return err
	}

	items, ok := payload[playlistTitle]
	if !ok {
		return fmt.Errorf("playlist %q not found", playlistTitle)
	}

	next := make([]BMPInfoItem, 0, len(items))
	removed := false
	for _, item := range items {
		if item.BVID == bvid {
			removed = true
			continue
		}
		next = append(next, item)
	}

	if !removed {
		return fmt.Errorf("bvid %s not found in playlist %q", bvid, playlistTitle)
	}

	if len(next) == 0 {
		delete(payload, playlistTitle)
	} else {
		payload[playlistTitle] = next
	}

	return b.writeBMPInfo(payload)
}

// EnsureTrackAudio refreshes one track when its cached audio is missing or expired.
func (b *Bili) EnsureTrackAudio(playlistTitle, bvid string) (BMPInfoItem, error) {
	if err := b.FixBMPInfo(bvid); err != nil {
		return BMPInfoItem{}, err
	}

	item, ok, err := b.FindTrack(playlistTitle, bvid)
	if err != nil {
		return BMPInfoItem{}, err
	}
	if !ok {
		return BMPInfoItem{}, fmt.Errorf("bvid %s not found in playlist %q", bvid, playlistTitle)
	}

	return item, nil
}

// RefreshTrackAudio forces one track to fetch a fresh audio URL and persists it.
func (b *Bili) RefreshTrackAudio(playlistTitle, bvid string) (BMPInfoItem, error) {
	payload, err := b.ReadBMPInfo()
	if err != nil {
		return BMPInfoItem{}, err
	}

	items, ok := payload[playlistTitle]
	if !ok {
		return BMPInfoItem{}, fmt.Errorf("playlist %q not found", playlistTitle)
	}

	for i := range items {
		if items[i].BVID != bvid {
			continue
		}

		if _, err := b.GetPlayInfo(items[i].BVID, PlayInfoPath); err != nil {
			return BMPInfoItem{}, err
		}

		audioURL, err := b.GetAudio()
		if err != nil {
			return BMPInfoItem{}, err
		}

		items[i].Audio = audioURL
		payload[playlistTitle] = items
		if err := b.writeBMPInfo(payload); err != nil {
			return BMPInfoItem{}, err
		}

		return items[i], nil
	}

	return BMPInfoItem{}, fmt.Errorf("bvid %s not found in playlist %q", bvid, playlistTitle)
}

// DownloadTrack refreshes one track if needed, then saves it under outputDir.
func (b *Bili) DownloadTrack(playlistTitle, bvid string, outputDir string) (string, error) {
	item, err := b.EnsureTrackAudio(playlistTitle, bvid)
	if err != nil {
		return "", err
	}

	fileName := sanitizeFileName(item.Title)
	if fileName == "" {
		fileName = item.BVID
	}
	filePath := filepath.Join(outputDir, fileName+".m4a")

	if err := b.AudioDownload(item.Audio, filePath); err != nil {
		return "", err
	}

	return filePath, nil
}

func sanitizeFileName(name string) string {
	clean := invalidFilenameChars.ReplaceAllString(name, "_")
	clean = strings.TrimSpace(clean)
	clean = strings.Trim(clean, ".")
	if clean == "" {
		return ""
	}
	return clean
}
