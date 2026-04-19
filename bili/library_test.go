package bili

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDeletePlaylist(t *testing.T) {
	withTempPaths(t)

	payload := BMPInfoPayload{
		"A": {{Title: "one", BVID: "BV1"}},
		"B": {{Title: "two", BVID: "BV2"}},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(BMPInfoPath, raw, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	client, err := New("")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := client.DeletePlaylist("A"); err != nil {
		t.Fatalf("DeletePlaylist() error = %v", err)
	}

	got, err := client.ReadBMPInfo()
	if err != nil {
		t.Fatalf("ReadBMPInfo() error = %v", err)
	}
	if _, exists := got["A"]; exists {
		t.Fatalf("playlist A still exists")
	}
	if _, exists := got["B"]; !exists {
		t.Fatalf("playlist B missing")
	}
}

func TestDeleteTrackDropsEmptyPlaylist(t *testing.T) {
	withTempPaths(t)

	payload := BMPInfoPayload{
		"A": {{Title: "one", BVID: "BV1"}},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(BMPInfoPath, raw, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	client, err := New("")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := client.DeleteTrack("A", "BV1"); err != nil {
		t.Fatalf("DeleteTrack() error = %v", err)
	}

	got, err := client.ReadBMPInfo()
	if err != nil {
		t.Fatalf("ReadBMPInfo() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty payload, got %#v", got)
	}
}
