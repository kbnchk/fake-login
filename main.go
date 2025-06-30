package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
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
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func main() {

	addr := os.Getenv("SERVER_ADDR")
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")

	if addr == "" {
		addr = ":8080"
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/api/login", loginHandler)

	slog.Info("Server running at " + addr)
	var serveErr error
	if certFile != "" && keyFile != "" {
		slog.Info("TLS enabled")
		serveErr = http.ListenAndServeTLS(addr, certFile, keyFile, nil)
	} else {
		serveErr = http.ListenAndServe(addr, nil)
	}
	if serveErr != nil {
		slog.Error(serveErr.Error())
	}
}
