package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"math/rand/v2"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

//go:embed static
var staticFS embed.FS

type asset struct {
	body  []byte
	ctype string
	etag  string
	html  bool
}

// clean URL -> файл в static/
var routes = map[string]string{
	"/":                "index.html",
	"/forgot-password": "forgot.html",
	"/privacy":         "privacy.html",
	"/terms":           "terms.html",
	"/style.css":       "style.css",
	"/script.js":       "script.js",
	"/favicon.svg":     "favicon.svg",
	"/favicon.ico":     "favicon.svg",
	"/robots.txt":      "robots.txt",
}

var (
	assets   = map[string]*asset{}
	notFound *asset
)

func loadAssets() error {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	load := func(name string) (*asset, error) {
		b, err := fs.ReadFile(sub, name)
		if err != nil {
			return nil, err
		}
		ct := mime.TypeByExtension(path.Ext(name))
		switch path.Ext(name) {
		case ".svg":
			ct = "image/svg+xml"
		case ".txt":
			ct = "text/plain; charset=utf-8"
		}
		sum := sha256.Sum256(b)
		return &asset{body: b, ctype: ct, etag: `"` + hex.EncodeToString(sum[:8]) + `"`, html: path.Ext(name) == ".html"}, nil
	}
	for url, name := range routes {
		a, err := load(name)
		if err != nil {
			return err
		}
		assets[url] = a
	}
	notFound, err = load("404.html")
	return err
}

func secure(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hd := w.Header()
		hd.Set("X-Content-Type-Options", "nosniff")
		hd.Set("X-Frame-Options", "DENY")
		hd.Set("Referrer-Policy", "no-referrer")
		hd.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		hd.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		h.ServeHTTP(w, r)
	})
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	a, ok := assets[r.URL.Path]
	status := http.StatusOK
	if !ok {
		a, status = notFound, http.StatusNotFound
	}
	hd := w.Header()
	hd.Set("Content-Type", a.ctype)
	if a.html {
		hd.Set("Cache-Control", "no-cache")
	} else {
		hd.Set("Cache-Control", "public, max-age=3600")
	}
	if status == http.StatusOK {
		hd.Set("ETag", a.etag)
		if r.Header.Get("If-None-Match") == a.etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = w.Write(a.body)
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// не даём накопить тысячи спящих обработчиков
var loginSlots = make(chan struct{}, 64)

func writeJSON(w http.ResponseWriter, status int, msg, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "message": msg})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSON(w, http.StatusMethodNotAllowed, "Method not allowed.", "method_not_allowed")
		return
	}
	select {
	case loginSlots <- struct{}{}:
		defer func() { <-loginSlots }()
	default:
		w.Header().Set("Retry-After", "30")
		writeJSON(w, http.StatusTooManyRequests, "Too many attempts. Please try again later.", "rate_limited")
		return
	}

	var creds loginRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, 4096))
	if err := dec.Decode(&creds); err != nil || creds.Username == "" || creds.Password == "" {
		writeJSON(w, http.StatusBadRequest, "Enter your email and password.", "invalid_request")
		return
	}
	// учётные данные никуда не пишутся и не логируются
	delay := time.Duration(700+rand.IntN(700)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-r.Context().Done():
		return
	}
	writeJSON(w, http.StatusUnauthorized, "Invalid email or password.", "invalid_credentials")
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

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}

func healthcheck(addr string, useTLS bool) int {
	port := addr[strings.LastIndex(addr, ":")+1:]
	scheme := "http"
	client := &http.Client{Timeout: 3 * time.Second}
	if useTLS {
		scheme = "https"
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	resp, err := client.Get(scheme + "://127.0.0.1:" + port + "/robots.txt")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func main() {
	addr := os.Getenv("SERVER_ADDR")
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if addr == "" {
		addr = ":8080"
	}
	useTLS := fileExists(certFile) && fileExists(keyFile)

	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		os.Exit(healthcheck(addr, useTLS))
	}
	if err := loadAssets(); err != nil {
		slog.Error("assets: " + err.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/", pageHandler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           secure(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	if useTLS {
		srv.TLSConfig = &tls.Config{MinVersion: parseTLSVersion(os.Getenv("TLS_MIN_VER"))}
	}

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	var err error
	if useTLS {
		slog.Info("Starting server at " + addr + " (TLS)")
		err = srv.ListenAndServeTLS(certFile, keyFile)
	} else {
		slog.Info("Starting server at " + addr)
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
