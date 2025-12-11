package brand

import (
	"encoding/json"
	"net/http"

	"service/access"
	"service/database"
	"service/log"
)

// just created a list for the dashboard but do optimized it pls
func init() {
	http.HandleFunc("/brand/list", func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET")
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodGet {
			header.Set("Content-Type", "application/json")

			uid, err := access.GetSessionUserID(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			allImgs, err := database.ListAllImages()
			if err != nil {
				log.Error("Failed to list images: %s", err.Error())
				http.Error(w, "Failed to list images", http.StatusInternalServerError)
				return
			}

			userImgs, err := database.FilterImagesByUser(allImgs, uid)
			if err != nil {
				log.Error("Failed to filter images: %s", err.Error())
				http.Error(w, "Failed to filter images", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(userImgs); err != nil {
				log.Error("Failed to encode response: %s", err.Error())
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
