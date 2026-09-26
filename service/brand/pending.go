package brand

import (
	"encoding/json"
	"net/http"
	"strconv"

	"service/access"
	"service/database"
	"service/discord"
	"service/log"
)

func init() {
	http.HandleFunc("/brand/pending", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", http.MethodGet)
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodGet {
			header.Set("Content-Type", "application/json")

			uidRes := access.GetSessionUserID(r)
			if uidRes.IsError() {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			uid := uidRes.MustGet()

			userRes := database.GetUser(uid)
			if userRes.IsError() {
				log.Error("Failed to get user: %s", userRes.Error())
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}
			u := userRes.MustGet()

			if !u.IsAdmin && !u.IsStaff {
				log.Error("User of ID %s is not admin or staff", u.ID)
				http.Error(w, "User is not admin or staff", http.StatusUnauthorized)
				return
			}

			imgListRes := database.ListPendingImages()
			if imgListRes.IsError() {
				log.Error("Failed to list pending images: %s", imgListRes.Error())
				http.Error(w, "Failed to list pending images", http.StatusInternalServerError)
				return
			}
			imgList := imgListRes.MustGet()

			query := r.URL.Query()
			userStr := query.Get("user")

			if userStr != "" {
				user, err := strconv.ParseUint(userStr, 10, 64)
				if err != nil {
					log.Error("Failed to get user ID: %s", err.Error())
					http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
					return
				}

				filteredImagesRes := database.FilterImagesByUser(imgList, user)
				if filteredImagesRes.IsError() {
					log.Error("Failed to filter images by user: %s", filteredImagesRes.Error())
					http.Error(w, "Failed to filter images", http.StatusInternalServerError)
					return
				}
				imgList = filteredImagesRes.MustGet()
			}

			for i, img := range imgList {
				userRes := database.GetUser(img.UserID)
				if userRes.IsError() {
					log.Error("Failed to get user for img %d: %s", img.ID, userRes.Error())
					continue
				}
				u := userRes.MustGet()
				imgList[i].Login = u.Login
			}

			log.Debug("Returning %d pending advertisements", len(imgList))

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(imgList); err != nil {
				log.Error("Failed to encode response: %s", err.Error())
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/brand/pending/accept", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", http.MethodPost)
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodPost {
			header.Set("Content-Type", "application/json")

			uidRes := access.GetSessionUserID(r)
			if uidRes.IsError() {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			uid := uidRes.MustGet()

			userRes := database.GetUser(uid)
			if userRes.IsError() {
				log.Error("Failed to get user: %s", userRes.Error())
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}
			u := userRes.MustGet()

			if !u.IsAdmin && !u.IsStaff {
				log.Error("User of ID %s is not admin or staff", u.ID)
				http.Error(w, "User is not admin or staff", http.StatusUnauthorized)
				return
			}

			query := r.URL.Query()
			idStr := query.Get("id")

			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				log.Error("Failed to get img ID: %s", err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			imgRes := database.ApproveImage(id)
			if imgRes.IsError() {
				log.Error("Failed to approve img: %s", imgRes.Error())
				http.Error(w, "Failed to approve img", http.StatusInternalServerError)
				return
			}
			img := imgRes.MustGet()

			webhookRes := discord.WebhookAccept(img, u)
			if webhookRes.IsError() {
				log.Warn(webhookRes.Error().Error())
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(img); err != nil {
				log.Error("Failed to encode response: %s", err.Error())
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
