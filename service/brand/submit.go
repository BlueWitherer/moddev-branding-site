package brand

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"service/access"
	"service/database"
	"service/log"
)

func init() {
	http.HandleFunc("/brand/submit", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "POST")
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodPost {
			header.Set("Content-Type", "application/json")

			uid, err := access.GetSessionUserID(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := database.GetUser(uid)
			if err != nil {
				log.Error("Failed to get ad owner: %s", err.Error())
				http.Error(w, "Failed to get ad owner", http.StatusInternalServerError)
				return
			}

			if user.Banned {
				log.Error("User %s is banned", user.Login)
				http.Error(w, "User is banned", http.StatusForbidden)
				return
			}

			// Parse form with 10MB limit
			r.ParseMultipartForm(10 << 20)

			// Get image file
			file, _, err := r.FormFile("image-upload")
			if err != nil {
				log.Error("Image not found: %s", err.Error())
				http.Error(w, "Image not found", http.StatusBadRequest)
				return
			}
			defer file.Close()

			// Create target folder
			targetDir := filepath.Join("..", "cdn")
			err = os.MkdirAll(targetDir, os.ModePerm)
			if err != nil {
				log.Error("Failed to get directory %s", err.Error())
				http.Error(w, "Failed to get directory", http.StatusInternalServerError)
				return
			}

			fileName := fmt.Sprintf("%d.webp", uid)
			dstPath := filepath.Join(targetDir, fileName)

			dst, err := os.Create(dstPath)
			if err != nil {
				log.Error("Failed to save image: %s", err.Error())
				http.Error(w, "Failed to save image", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				log.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			imageURL := fmt.Sprintf("%s/cdn/%s", access.GetDomain(r), fileName)
			imgID, err := database.CreateImage(uid, imageURL)
			if err != nil {
				e := os.Remove(dstPath)
				if e != nil {
					log.Error("Failed to delete brand image: %s", e.Error())
				}

				log.Error("Failed to create brand image row: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if user.IsAdmin || user.IsStaff || user.Verified {
				newImg, err := database.ApproveImage(imgID)
				if err != nil {
					log.Error("Failed to auto-approve new img by verified user: %s", err.Error())
				} else {
					log.Info("Auto-approved img %s (%v) by verified user %s (%s)", newImg.ImageURL, newImg.ID, user.Login, user.ID)
				}
			}

			var out struct {
				ID       uint64 `json:"id"`
				ImageURL string `json:"image_url"`
			}

			out.ID = imgID
			out.ImageURL = imageURL

			log.Info("Saved img to %s, id=%v, user_id=%s", dstPath, imgID, uid)

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(out); err != nil {
				log.Error("Failed to encode response: %s", err.Error())
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
