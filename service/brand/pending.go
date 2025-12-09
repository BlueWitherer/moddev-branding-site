package brand

import (
	"encoding/json"
	"net/http"
	"strconv"

	"service/access"
	"service/database"
	"service/log"
)

func init() {
	http.HandleFunc("/ads/pending", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET")
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodGet {
			header.Set("Content-Type", "application/json")

			// require login
			uid, err := access.GetSessionUserID(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			u, err := database.GetUser(uid)
			if err != nil {
				log.Error("Failed to get user: %s", err.Error())
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}

			if !u.IsAdmin && !u.IsStaff {
				log.Error("User of ID %s is not admin or staff", u.ID)
				http.Error(w, "User is not admin or staff", http.StatusUnauthorized)
				return
			}

			// Get pending ads directly from database with WHERE pending != 0
			adList, err := database.ListPendingImages()
			if err != nil {
				log.Error("Failed to list pending ads: %s", err.Error())
				http.Error(w, "Failed to list pending ads", http.StatusInternalServerError)
				return
			}

			query := r.URL.Query()
			userStr := query.Get("user")

			user, err := strconv.ParseUint(userStr, 10, 64)
			if err != nil {
				log.Error("Failed to get user ID: %s", err.Error())
				http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
				return
			}

			adList, err = database.FilterImagesByUser(adList, user)
			if err != nil {
				log.Error("Failed to filter ads by user: %s", err.Error())
				http.Error(w, "Failed to filter ads", http.StatusInternalServerError)
				return
			}

			log.Debug("Returning %d pending advertisements", len(adList))

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(adList); err != nil {
				log.Error("Failed to encode response: %s", err.Error())
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/ads/pending/accept", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "POST")
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodPost {
			header.Set("Content-Type", "application/json")

			// require login
			uid, err := access.GetSessionUserID(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			u, err := database.GetUser(uid)
			if err != nil {
				log.Error("Failed to get user: %s", err.Error())
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}

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

			img, err := database.ApproveImage(id)
			if err != nil {
				log.Error("Failed to approve img: %s", err.Error())
				http.Error(w, "Failed to approve img", http.StatusInternalServerError)
				return
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
