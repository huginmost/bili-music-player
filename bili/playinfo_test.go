package bili

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPlayInfoPatternExtractsJSON(t *testing.T) {
	html := `
	<html>
		<body>
			<script>
				window.__playinfo__ = {
					"data": {
						"dash": {
							"audio": [{"id": 1, "baseUrl": "https://example.com/audio.m4a"}]
						}
					}
				}
			</script>
		</body>
	</html>`

	matches := playInfoPattern.FindStringSubmatch(html)
	if len(matches) < 2 {
		t.Fatalf("expected playinfo JSON to be extracted")
	}

	if matches[1] == "" {
		t.Fatalf("expected non-empty JSON match")
	}
}

func TestInitialStatePatternExtractsJSON(t *testing.T) {
	html := `
	<html>
		<body>
			<script>
				window.__INITIAL_STATE__ = {
					"bvid": "BV1oU1jBXEN8",
					"videoData": {
						"title": "test-title"
					}
				};(function() {
					console.log("next");
				})();
			</script>
		</body>
	</html>`

	matches := initialStatePattern.FindStringSubmatch(html)
	if len(matches) < 2 {
		t.Fatalf("expected initial state JSON to be extracted")
	}

	if matches[1] == "" {
		t.Fatalf("expected non-empty JSON match")
	}
}

func TestListPlayInfoPatternExtractsJSON(t *testing.T) {
	html := `
	<html>
		<body>
			<script>
				window.__playinfo__ = {
					"mediaListInfo": {
						"title": "list-title"
					}
				}
			</script>
		</body>
	</html>`

	matches := playInfoPattern.FindStringSubmatch(html)
	if len(matches) < 2 {
		t.Fatalf("expected playinfo JSON to be extracted from list page")
	}

	if matches[1] == "" {
		t.Fatalf("expected non-empty JSON match")
	}
}

func TestBiliJSParsesAndFormatsJSON(t *testing.T) {
	client, err := New("")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	tempDir := filepath.Join("..", ".testdata")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	tempFile := filepath.Join(tempDir, "pi-"+time.Now().Format("20060102150405")+".json")
	content := `{"data":{"dash":{"audio":[{"id":1,"baseUrl":"https://example.com/audio.m4a"}]}}}`
	if err := os.WriteFile(tempFile, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	defer os.Remove(tempFile)

	parsed, formatted, err := client.ParseJSON(tempFile)
	if err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}

	if parsed["data"] == nil {
		t.Fatalf("expected parsed data field")
	}

	if len(formatted) == 0 || formatted[0] != '{' {
		t.Fatalf("expected formatted JSON output, got %q", formatted)
	}
}
