package bili

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// AudioDownload downloads the binary content from url to filePath.
// The request includes the Bilibili origin header to avoid access denial.
func (b *Bili) AudioDownload(url, filePath string) error {
	if b == nil {
		return fmt.Errorf("bili is nil")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Origin", baseURL[:len(baseURL)-1])
	req.Header.Set("Referer", baseURL)

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	outputDir := filepath.Dir(filePath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return err
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}

	return nil
}
