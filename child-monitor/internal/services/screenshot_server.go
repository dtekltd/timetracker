package services

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"child-monitor/internal/logger"
)

var screenshotServerURL string

// StartScreenshotServer starts a local HTTP file server that serves screenshot
// files from the given root directory. It binds to 127.0.0.1 on a random free
// port and returns the base URL.
func StartScreenshotServer(screenshotRoot string) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("listen for screenshot server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	screenshotServerURL = baseURL

	mux := http.NewServeMux()
	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}

		// Reject path traversal attempts.
		if strings.Contains(path, "..") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Ensure the requested file is within screenshotRoot.
		if !strings.HasPrefix(path, screenshotRoot) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "http://wails.localhost")
		http.ServeFile(w, r, path)
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("screenshot server error", err)
		}
	}()

	logger.Info(fmt.Sprintf("screenshot server started at %s", baseURL))
	return baseURL, nil
}

// GetScreenshotServerURL returns the base URL of the running screenshot server.
func GetScreenshotServerURL() string {
	return screenshotServerURL
}
