package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/huginmost/bili-music-player/bili"
)

type app struct {
	client *bili.Bili
}

type trackRequest struct {
	PlaylistTitle string `json:"playlistTitle"`
	BVID          string `json:"bvid"`
}

type prefetchRequest struct {
	PlaylistTitle string   `json:"playlistTitle"`
	BVIDs         []string `json:"bvids"`
}

type downloadResponse struct {
	FilePath string `json:"filePath"`
}

func main() {
	client, err := bili.New(os.Getenv("BILI_COOKIE"))
	if err != nil {
		log.Fatalf("bili_init failed: %v", err)
	}

	server := &app{client: client}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", server.handleHealth)
	mux.HandleFunc("/api/library", server.handleLibrary)
	mux.HandleFunc("/api/tracks/refresh", server.handleRefreshTrack)
	mux.HandleFunc("/api/tracks/prefetch", server.handlePrefetchTracks)
	mux.HandleFunc("/api/playlists", server.handleDeletePlaylist)
	mux.HandleFunc("/api/tracks", server.handleDeleteTrack)
	mux.HandleFunc("/api/downloads", server.handleDownloadTrack)
	mux.HandleFunc("/media/cover", server.handleCoverProxy)
	mux.HandleFunc("/media/audio", server.handleAudioProxy)

	if staticDir, ok := resolveStaticDir(); ok {
		mux.Handle("/", newStaticHandler(staticDir))
	}

	addr := ":8765"
	log.Printf("bmplayer web server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"reachable": a.client.Try(),
	})
}

func (a *app) handleLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := a.client.ReadBMPInfo()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (a *app) handleRefreshTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req trackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	item, err := a.client.EnsureTrackAudio(req.PlaylistTitle, req.BVID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (a *app) handlePrefetchTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req prefetchRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	limit := len(req.BVIDs)
	if limit > 3 {
		limit = 3
	}

	refreshed := make([]bili.BMPInfoItem, 0, limit)
	for _, bvid := range req.BVIDs[:limit] {
		if strings.TrimSpace(bvid) == "" {
			continue
		}
		item, err := a.client.EnsureTrackAudio(req.PlaylistTitle, bvid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		refreshed = append(refreshed, item)
	}

	writeJSON(w, http.StatusOK, refreshed)
}

func (a *app) handleDeletePlaylist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := a.client.DeletePlaylist(req.Title); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *app) handleDeleteTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req trackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := a.client.DeleteTrack(req.PlaylistTitle, req.BVID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *app) handleDownloadTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req trackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	filePath, err := a.client.DownloadTrack(req.PlaylistTitle, req.BVID, "downloads")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, downloadResponse{FilePath: filePath})
}

func (a *app) handleCoverProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	src := strings.TrimSpace(r.URL.Query().Get("src"))
	if src == "" {
		http.Error(w, "missing src", http.StatusBadRequest)
		return
	}

	if err := a.client.ProxyImage(w, r, src); err != nil {
		writeError(w, http.StatusBadGateway, err)
	}
}

func (a *app) handleAudioProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playlistTitle := strings.TrimSpace(r.URL.Query().Get("playlistTitle"))
	bvid := strings.TrimSpace(r.URL.Query().Get("bvid"))
	if playlistTitle == "" || bvid == "" {
		http.Error(w, "missing playlistTitle or bvid", http.StatusBadRequest)
		return
	}

	if err := a.client.ProxyTrackAudio(w, r, playlistTitle, bvid); err != nil {
		writeError(w, http.StatusBadGateway, err)
	}
}

func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": fmt.Sprintf("%v", err),
	})
}

func resolveStaticDir() (string, bool) {
	exePath, err := os.Executable()
	if err != nil {
		return "", false
	}

	exeDir := filepath.Dir(exePath)
	cwd, _ := os.Getwd()

	candidates := []string{
		filepath.Join(exeDir, "frontend"),
		filepath.Join(exeDir, "frontend", "dist"),
		filepath.Join(cwd, "frontend"),
		filepath.Join(cwd, "frontend", "dist"),
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil || !info.IsDir() {
			continue
		}

		indexPath := filepath.Join(candidate, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return candidate, true
		}
	}

	return "", false
}

func newStaticHandler(root string) http.Handler {
	fileServer := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}

		relativePath := strings.TrimPrefix(filepath.Clean(r.URL.Path), string(filepath.Separator))
		relativePath = strings.TrimPrefix(relativePath, "/")
		target := filepath.Join(root, relativePath)
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
}
