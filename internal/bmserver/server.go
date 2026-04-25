package bmserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/huginmost/bili-music-player/bili"
)

type Server struct {
	client *bili.Bili
	mux    *http.ServeMux
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

type importRequest struct {
	ID string `json:"id"`
}

func NewFromEnv() (*Server, error) {
	client, err := bili.New(os.Getenv("BILI_COOKIE"))
	if err != nil {
		return nil, fmt.Errorf("bili_init failed: %w", err)
	}

	server := &Server{
		client: client,
		mux:    http.NewServeMux(),
	}
	server.routes()

	return server, nil
}

func (s *Server) Handler() http.Handler {
	return withCORS(s.mux)
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("bmplayer web server listening on http://localhost%s", addr)
	return http.ListenAndServe(addr, s.Handler())
}

func (s *Server) StartLocalhost() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	addr := listener.Addr().String()
	go func() {
		log.Printf("bmplayer desktop api listening on http://%s", addr)
		if err := http.Serve(listener, s.Handler()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("bmplayer desktop api stopped: %v", err)
		}
	}()

	return "http://" + addr, nil
}

func (s *Server) EnableStaticFiles() {
	if staticDir, ok := resolveStaticDir(); ok {
		s.mux.Handle("/", newStaticHandler(staticDir))
	}
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/library", s.handleLibrary)
	s.mux.HandleFunc("/api/settings", s.handleSettings)
	s.mux.HandleFunc("/api/tracks/refresh", s.handleRefreshTrack)
	s.mux.HandleFunc("/api/tracks/prefetch", s.handlePrefetchTracks)
	s.mux.HandleFunc("/api/library/import/video", s.handleImportVideo)
	s.mux.HandleFunc("/api/library/import/list", s.handleImportList)
	s.mux.HandleFunc("/api/playlists", s.handleDeletePlaylist)
	s.mux.HandleFunc("/api/tracks", s.handleDeleteTrack)
	s.mux.HandleFunc("/api/downloads", s.handleDownloadTrack)
	s.mux.HandleFunc("/media/cover", s.handleCoverProxy)
	s.mux.HandleFunc("/media/audio", s.handleAudioProxy)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"reachable": s.client.Try(),
	})
}

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := s.client.ReadBMPInfo()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := s.client.ReadSettings()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, settings)
	case http.MethodPut:
		var settings bili.PlayerSettings
		if err := decodeJSON(r, &settings); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if err := s.client.WriteSettings(settings); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleRefreshTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req trackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	item, err := s.client.EnsureTrackAudio(req.PlaylistTitle, req.BVID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handlePrefetchTracks(w http.ResponseWriter, r *http.Request) {
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
		item, err := s.client.EnsureTrackAudio(req.PlaylistTitle, bvid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		refreshed = append(refreshed, item)
	}

	writeJSON(w, http.StatusOK, refreshed)
}

func (s *Server) handleImportVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req importRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	bvid := strings.TrimSpace(req.ID)
	if bvid == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing bv id"))
		return
	}

	if _, err := s.client.GetPlayInfo(bvid, bili.PlayInfoPath); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("bili_get_pi failed: %w", err))
		return
	}
	if _, err := s.client.GetInitialState(bvid, bili.InitialStatePath); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("bili_get_is failed: %w", err))
		return
	}
	if _, err := s.client.GetBMPInfo(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("bili_get_bmpinfo failed: %w", err))
		return
	}

	payload, err := s.client.ReadBMPInfo()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleImportList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req importRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	listID := strings.TrimSpace(req.ID)
	if listID == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing list id"))
		return
	}

	if _, err := s.client.GetListInitialState(listID, bili.InitialStatePath); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("bili_lget_is failed: %w", err))
		return
	}
	if _, err := s.client.GetListBMPInfo(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("bili_lget_bmpinfo failed: %w", err))
		return
	}

	payload, err := s.client.ReadBMPInfo()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleDeletePlaylist(w http.ResponseWriter, r *http.Request) {
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

	if err := s.client.DeletePlaylist(req.Title); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req trackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.client.DeleteTrack(req.PlaylistTitle, req.BVID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDownloadTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req trackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	filePath, err := s.client.DownloadTrack(req.PlaylistTitle, req.BVID, "downloads")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, downloadResponse{FilePath: filePath})
}

func (s *Server) handleCoverProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	src := strings.TrimSpace(r.URL.Query().Get("src"))
	if src == "" {
		http.Error(w, "missing src", http.StatusBadRequest)
		return
	}

	if err := s.client.ProxyImage(w, r, src); err != nil {
		writeError(w, http.StatusBadGateway, err)
	}
}

func (s *Server) handleAudioProxy(w http.ResponseWriter, r *http.Request) {
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

	if err := s.client.ProxyTrackAudio(w, r, playlistTitle, bvid); err != nil {
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
