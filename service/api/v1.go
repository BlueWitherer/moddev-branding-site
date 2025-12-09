package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"service/database"
	"service/log"
)

func init() {
	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Mod Developer Branding API v1 service pinged")
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET")
		header.Set("Access-Control-Allow-Headers", "Content-Type")
		header.Set("Content-Type", "text/plain")

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong!")
	})

	http.HandleFunc("/api/v1/image", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Getting developer branding image...")
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET")
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodGet {
			header.Set("Content-Type", "image/webp")

			query := r.URL.Query()
			dev := query.Get("dev")

			user, err := database.GetUserFromLogin(dev)
			if err != nil {
				log.Error("Failed to get user: %s", err.Error())
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}

			img, err := database.GetImageForUser(user.ID)
			if err != nil {
				log.Error("Failed to get image info: %s", err.Error())
				http.Error(w, "Failed to get image info", http.StatusInternalServerError)
				return
			}

			if img.Pending {
				log.Error("Image still pending review")
				http.Error(w, "Image still pending review", http.StatusForbidden)
				return
			}

			fileName := fmt.Sprintf("%d.webp", user.ID)
			dstPath := filepath.Join(filepath.Join("..", "cdn"), fileName)

			f, err := os.Open(dstPath)
			if err != nil {
				log.Error("Failed to open image: %s", err.Error())
				http.Error(w, "Failed to open image", http.StatusNotFound)
				return
			}
			defer f.Close()

			w.WriteHeader(http.StatusOK)
			if _, err := io.Copy(w, f); err != nil {
				log.Error("Failed to stream image: %s", err.Error())
				http.Error(w, "Failed to stream image", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
