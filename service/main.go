package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"service/access"
	_ "service/api"
	_ "service/brand"
	"service/log"

	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

var visitors = cache.New(15*time.Minute, 30*time.Minute)

func getClientIP(r *http.Request) string {
	if cf := r.Header.Get("CF-Connecting-IP"); cf != "" {
		return cf
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func getVisitor(ip string) *rate.Limiter {
	if val, found := visitors.Get(ip); found {
		return val.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(10, 30)
	visitors.Set(ip, limiter, cache.DefaultExpiration)
	return limiter
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := getVisitor(ip)

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Print("Starting server...")

	port := os.Getenv("WEB_PORT")
	if port == "" {
		log.Error("WEB_PORT is not set")
	}

	srv := &http.Server{Addr: fmt.Sprintf(":%s", port)}

	// SPA fallback
	log.Debug("Setting up SPA fallback for client-side routing")
	staticDir := "../dist"
	fs := http.FileServer(http.Dir(staticDir))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request for host %s", access.FullURL(r))

		requestedPath := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		fullPath := filepath.Join(staticDir, requestedPath)
		if requestedPath == "" || requestedPath == "." {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}

		log.Debug("Serving index.html for SPA route: %s", r.URL.Path)
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	log.Debug("Starting image handler...")
	http.HandleFunc("/cdn/", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		requestedPath := strings.TrimPrefix(r.URL.Path, "/cdn/")
		fullPath := filepath.Join("..", "cdn", requestedPath)

		header.Set("Content-Type", "image/webp")

		http.ServeFile(w, r, fullPath)
	})

	log.Debug("Starting handlers...")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Recovered from panic: %v", r)
			}
		}()

		log.Done("Server started successfully on host http://localhost%s", srv.Addr)
		srv.Handler = rateLimitMiddleware(http.DefaultServeMux)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err.Error())
		}
	}()

	// shutdown sequence
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Warn("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Shutdown error: %s", err.Error())
	} else {
		log.Print("Server stopped")
	}
}
