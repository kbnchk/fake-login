package main

import (
	"crypto/tls"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds LoginRequest
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil || creds.Username == "" || creds.Password == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	time.Sleep(1 * time.Second)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func parseTLSVersion(s string) uint16 {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "1.0", "TLS1":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	default:
		return tls.VersionTLS13
	}
}

func main() {

	addr := os.Getenv("SERVER_ADDR")
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	minTlsVerEnv := os.Getenv("TLS_MIN_VER")

	if addr == "" {
		addr = ":8080"
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/login", loginHandler)

	_, certErr := os.Stat(certFile)
	_, keyErr := os.Stat(keyFile)

	slog.Info("Starting server at " + addr)
	var serveErr error
	if certErr == nil && keyErr == nil {
		minTls := parseTLSVersion(minTlsVerEnv)
		slog.Info("TLS enabled")
		server := &http.Server{
			Addr: addr,
			TLSConfig: &tls.Config{
				MinVersion: minTls,
			},
		}
		serveErr = server.ListenAndServeTLS(certFile, keyFile)

	} else {
		slog.Info("No SSL certificate provided")
		serveErr = http.ListenAndServe(addr, nil)
	}
	if serveErr != nil {
		slog.Error(serveErr.Error())
	}
}
