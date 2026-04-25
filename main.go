package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/huginmost/bili-music-player/internal/bmserver"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dist, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatalf("failed to load frontend assets: %v", err)
	}

	server, err := bmserver.NewFromEnv()
	if err != nil {
		log.Fatalf("failed to create desktop api: %v", err)
	}

	app := NewApp()

	err = wails.Run(&options.App{
		Title:            "BMPlayer",
		Width:            1180,
		Height:           760,
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets:     dist,
			Handler:    server.Handler(),
			Middleware: injectAPIBase(dist, "/api"),
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windowsoptions.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}

func injectAPIBase(assets fs.FS, apiBase string) assetserver.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && (r.URL.Path == "/" || r.URL.Path == "/index.html") {
				indexHTML, err := fs.ReadFile(assets, "index.html")
				if err == nil {
					injected := strings.Replace(
						string(indexHTML),
						"<head>",
						fmt.Sprintf("<head><script>window.BMPLAYER_API_BASE=%q;</script>", apiBase),
						1,
					)
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					_, _ = w.Write([]byte(injected))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
