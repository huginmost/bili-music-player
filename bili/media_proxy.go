package bili

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ProxyImage streams one remote image through the local server.
func (b *Bili) ProxyImage(w http.ResponseWriter, r *http.Request, mediaURL string) error {
	headers := map[string]string{
		"User-Agent": defaultUserAgent,
		"Referer":    baseURL,
	}

	return b.proxyMedia(w, r, mediaURL, headers)
}

// ProxyTrackAudio refreshes the track when needed and then streams audio through the local server.
func (b *Bili) ProxyTrackAudio(w http.ResponseWriter, r *http.Request, playlistTitle, bvid string) error {
	item, err := b.EnsureTrackAudio(playlistTitle, bvid)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"User-Agent": defaultUserAgent,
		"Referer":    baseURL,
		"Origin":     "https://www.bilibili.com",
	}
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		headers["Range"] = rangeHeader
	}

	if err := b.proxyMedia(w, r, item.Audio, headers); err != nil {
		if !shouldRetryProxyAudio(err) {
			return err
		}

		item, refreshErr := b.RefreshTrackAudio(playlistTitle, bvid)
		if refreshErr != nil {
			return refreshErr
		}

		return b.proxyMedia(w, r, item.Audio, headers)
	}

	return nil
}

func (b *Bili) proxyMedia(w http.ResponseWriter, r *http.Request, mediaURL string, headers map[string]string) error {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, mediaURL, nil)
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	copyProxyHeader(w.Header(), resp.Header, "Accept-Ranges")
	copyProxyHeader(w.Header(), resp.Header, "Cache-Control")
	copyProxyHeader(w.Header(), resp.Header, "Content-Length")
	copyProxyHeader(w.Header(), resp.Header, "Content-Range")
	copyProxyHeader(w.Header(), resp.Header, "Content-Type")
	copyProxyHeader(w.Header(), resp.Header, "ETag")
	copyProxyHeader(w.Header(), resp.Header, "Last-Modified")

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

func copyProxyHeader(dst, src http.Header, key string) {
	if value := src.Get(key); value != "" {
		dst.Set(key, value)
	}
}

func shouldRetryProxyAudio(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return strings.Contains(message, "unexpected status code: 401") ||
		strings.Contains(message, "unexpected status code: 403") ||
		strings.Contains(message, "unexpected status code: 404") ||
		strings.Contains(message, "unexpected status code: 410")
}
