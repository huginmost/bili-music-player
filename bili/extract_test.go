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
	group, ok := payload["playlist - BV1"]
	if !ok {
		t.Fatalf("expected merged playlist key in payload: %+v", payload)
	}
	if len(group) != 2 {
		t.Fatalf("expected 2 payload items, got %d", len(group))
	}
	if group[1].Audio != "" {
		t.Fatalf("expected empty audio field, got %q", group[1].Audio)
	}
}

func TestGetBMPInfoKeepsExistingGroups(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	existing := BMPInfoPayload{
		"old-playlist - OLD": []BMPInfoItem{
			{Title: "old-song", Pic: "old-pic", BVID: "OLD", Audio: ""},
		},
	}
	raw, err := json.Marshal(existing)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if err := os.WriteFile(BMPInfoPath, raw, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	content := `{"ugc_season":{"title":"new-playlist","sections":[{"episodes":[{"title":"song-1","arc_pic":"pic1","bvid":"BV1"}]}]}}`
	if err := os.WriteFile(InitialStatePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if _, err := client.GetBMPInfo(); err != nil {
		t.Fatalf("GetBMPInfo returned error: %v", err)
	}

	payload, err := client.ReadBMPInfo()
	if err != nil {
		t.Fatalf("ReadBMPInfo returned error: %v", err)
	}
	if len(payload) != 2 {
		t.Fatalf("expected 2 playlist groups, got %d", len(payload))
	}
	if payload["old-playlist - OLD"][0].BVID != "OLD" {
		t.Fatalf("expected old playlist to be preserved, got %+v", payload["old-playlist"])
	}
	if payload["new-playlist - BV1"][0].BVID != "BV1" {
		t.Fatalf("expected new playlist to be added, got %+v", payload["new-playlist - BV1"])
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

func TestGetListTitleReturnsNestedTitle(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	content := `{"mediaListInfo":{"title":"my-list-title"}}`
	if err := os.WriteFile(InitialStatePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	title := client.GetListTitle()
	if title != "my-list-title" {
		t.Fatalf("expected list title %q, got %q", "my-list-title", title)
	}
}

func TestGetListTitleReturnsEmptyWhenMissing(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	content := `{"data":{"dash":{}}}`
	if err := os.WriteFile(InitialStatePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	title := client.GetListTitle()
	if title != "" {
		t.Fatalf("expected empty list title, got %q", title)
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
		"playlist": []BMPInfoItem{
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
	if got["playlist"][0].Audio != "https://example.com/a.m4a" {
		t.Fatalf("expected audio field to round-trip, got %q", got["playlist"][0].Audio)
	}
}

func TestReadBMPInfoAllowsEmptyFile(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	if err := os.WriteFile(BMPInfoPath, []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := client.ReadBMPInfo()
	if err != nil {
		t.Fatalf("ReadBMPInfo returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty payload, got %+v", got)
	}
}

func TestGetListBMPInfoWritesGroupedList(t *testing.T) {
	client := newTestClient(t)
	withTempPaths(t)

	initialStateContent := `{"mediaListInfo":{"title":"list-title"},"playlist":{"id":12345},"resourceList":[{"title":"song-1","cover":"cover-1","bvid":"BV1"},{"title":"song-2","cover":"cover-2","bvid":"BV2"}]}`
	if err := os.WriteFile(InitialStatePath, []byte(initialStateContent), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	items, err := client.GetListBMPInfo()
	if err != nil {
		t.Fatalf("GetListBMPInfo returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	payload, err := client.ReadBMPInfo()
	if err != nil {
		t.Fatalf("ReadBMPInfo returned error: %v", err)
	}
	if payload["list-title - 12345"][0].BVID != "BV1" {
		t.Fatalf("expected list items to store under merged key, got %+v", payload["list-title - 12345"])
	}
}
