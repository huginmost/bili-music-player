package bili

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetUGCSeasonTitleFallback(t *testing.T) {
	client, err := New("")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	dir := filepath.Join("..", ".testdata", "extract-"+time.Now().Format("20060102150405")+"-title")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	oldISPath := InitialStatePath
	InitialStatePath = filepath.Join(dir, "is.json")
	defer func() { InitialStatePath = oldISPath }()

	content := `{"ugc_season":{"id":1,"title":"❅·°","cover":"x"}}`
	if err := os.WriteFile(InitialStatePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	defer os.Remove(InitialStatePath)

	title, err := client.GetUGCSeasonTitle()
	if err != nil {
		t.Fatalf("GetUGCSeasonTitle returned error: %v", err)
	}
	if title != "❅·°" {
		t.Fatalf("expected title %q, got %q", "❅·°", title)
	}
}

func TestGetBMPInfoWritesPlaylist(t *testing.T) {
	client, err := New("")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	dir := filepath.Join("..", ".testdata", "extract-"+time.Now().Format("20060102150405")+"-bmp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	oldISPath := InitialStatePath
	oldBMPInfoPath := BMPInfoPath
	InitialStatePath = filepath.Join(dir, "is.json")
	BMPInfoPath = filepath.Join(dir, "bmpinfo.json")
	defer func() {
		InitialStatePath = oldISPath
		BMPInfoPath = oldBMPInfoPath
	}()

	content := `{"ugc_season":{"title":"合集","sections":[{"episodes":[{"title":"歌1","arc_pic":"pic1","bvid":"BV1"},{"title":"歌2","pic":"pic2","bvid":"BV2"}]}]}}`
	if err := os.WriteFile(InitialStatePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	defer os.Remove(InitialStatePath)
	defer os.Remove(BMPInfoPath)

	items, err := client.GetBMPInfo()
	if err != nil {
		t.Fatalf("GetBMPInfo returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "歌1" || items[0].Pic != "pic1" || items[0].BVID != "BV1" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if _, err := os.Stat(BMPInfoPath); err != nil {
		t.Fatalf("expected bmpinfo file to exist: %v", err)
	}
}

func TestGetAudioReturnsFirstBaseURL(t *testing.T) {
	client, err := New("")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	dir := filepath.Join("..", ".testdata", "extract-"+time.Now().Format("20060102150405")+"-audio")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	oldPIPath := PlayInfoPath
	PlayInfoPath = filepath.Join(dir, "pi.json")
	defer func() { PlayInfoPath = oldPIPath }()

	content := `{"data":{"dash":{"audio":[{"baseUrl":"https://example.com/audio.m4a"}]}}}`
	if err := os.WriteFile(PlayInfoPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	defer os.Remove(PlayInfoPath)

	audioURL, err := client.GetAudio()
	if err != nil {
		t.Fatalf("GetAudio returned error: %v", err)
	}
	if audioURL != "https://example.com/audio.m4a" {
		t.Fatalf("expected audio URL, got %q", audioURL)
	}
}
