package bili

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestClient(t *testing.T) *Bili {
	t.Helper()

	client, err := New("")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	return client
}

func withTempPaths(t *testing.T) string {
	t.Helper()

	dir := filepath.Join("..", ".testdata", "extract-"+time.Now().Format("20060102150405.000000000"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	oldPIPath := PlayInfoPath
	oldISPath := InitialStatePath
	oldBMPInfoPath := BMPInfoPath

	PlayInfoPath = filepath.Join(dir, "pi.json")
	InitialStatePath = filepath.Join(dir, "is.json")
	BMPInfoPath = filepath.Join(dir, "bmpinfo.json")

	t.Cleanup(func() {
		PlayInfoPath = oldPIPath
		InitialStatePath = oldISPath
		BMPInfoPath = oldBMPInfoPath
		_ = os.RemoveAll(dir)
	})

	return dir
}

func TestGetUGCSeasonTitleFallback(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	content := `{"ugc_season":{"id":1,"title":"season-title","cover":"x"}}`
	if err := os.WriteFile(InitialStatePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	title, err := client.GetUGCSeasonTitle()
	if err != nil {
		t.Fatalf("GetUGCSeasonTitle returned error: %v", err)
	}
	if title != "season-title" {
		t.Fatalf("expected title %q, got %q", "season-title", title)
	}
}

func TestGetBMPInfoWritesPlaylist(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	content := `{"ugc_season":{"title":"playlist","sections":[{"episodes":[{"title":"song-1","arc_pic":"pic1","bvid":"BV1"},{"title":"song-2","pic":"pic2","bvid":"BV2"}]}]}}`
	if err := os.WriteFile(InitialStatePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	items, err := client.GetBMPInfo()
	if err != nil {
		t.Fatalf("GetBMPInfo returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "song-1" || items[0].Pic != "pic1" || items[0].BVID != "BV1" || items[0].Audio != "" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}

	payload, err := client.ReadBMPInfo()
	if err != nil {
		t.Fatalf("ReadBMPInfo returned error: %v", err)
	}
	if len(payload.PLInfo) != 2 {
		t.Fatalf("expected 2 payload items, got %d", len(payload.PLInfo))
	}
	if payload.PLInfo[1].Audio != "" {
		t.Fatalf("expected empty audio field, got %q", payload.PLInfo[1].Audio)
	}
}

func TestGetAudioReturnsHighestBandwidthBaseURL(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	content := `{"data":{"dash":{"audio":[{"baseUrl":"https://example.com/low.m4a","bandwidth":100},{"baseUrl":"https://example.com/high.m4a","bandwidth":200}]}}}`
	if err := os.WriteFile(PlayInfoPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	audioURL, err := client.GetAudio()
	if err != nil {
		t.Fatalf("GetAudio returned error: %v", err)
	}
	if audioURL != "https://example.com/high.m4a" {
		t.Fatalf("expected audio URL, got %q", audioURL)
	}
}

func TestAudioDownloadWritesFile(t *testing.T) {
	client := newTestClient(t)
	dir := withTempPaths(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Origin"); got != "https://www.bilibili.com" {
			t.Fatalf("expected Origin header, got %q", got)
		}
		if got := r.Header.Get("Referer"); got != "https://www.bilibili.com/" {
			t.Fatalf("expected Referer header, got %q", got)
		}
		_, _ = w.Write([]byte("audio-data"))
	}))
	defer server.Close()

	outputPath := filepath.Join(dir, "test.m4a")
	if err := client.AudioDownload(server.URL, outputPath); err != nil {
		t.Fatalf("AudioDownload returned error: %v", err)
	}

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(got) != "audio-data" {
		t.Fatalf("expected downloaded content, got %q", string(got))
	}
}

func TestReadBMPInfoPreservesAudioField(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	payload := BMPInfoPayload{
		PLInfo: []BMPInfoItem{
			{Title: "song-1", Pic: "pic1", BVID: "BV1", Audio: "https://example.com/a.m4a"},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if err := os.WriteFile(BMPInfoPath, raw, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := client.ReadBMPInfo()
	if err != nil {
		t.Fatalf("ReadBMPInfo returned error: %v", err)
	}
	if got.PLInfo[0].Audio != "https://example.com/a.m4a" {
		t.Fatalf("expected audio field to round-trip, got %q", got.PLInfo[0].Audio)
	}
}
