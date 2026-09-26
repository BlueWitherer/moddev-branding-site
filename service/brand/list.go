package brand

import (
	"encoding/json"
	"net/http"

	"service/access"
	"service/database"
	"service/log"
)

func init() {
	http.HandleFunc("/brand/list", func(w http.ResponseWriter, r *http.Request) {
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

			imgListRes := database.ListAllImages()
			if imgListRes.IsError() {
				log.Error("Failed to list images: %s", imgListRes.Error())
				http.Error(w, "Failed to list images", http.StatusInternalServerError)
				return
			}
			imgList := imgListRes.MustGet()

			userImagesRes := database.FilterImagesByUser(imgList, uid)
			if userImagesRes.IsError() {
				log.Error("Failed to filter images for user %d: %s", uid, userImagesRes.Error())
				http.Error(w, "Failed to filter images", http.StatusInternalServerError)
				return
			}
			userImages := userImagesRes.MustGet()

			log.Debug("Returning %d images for user %d", len(userImages), uid)

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(userImages); err != nil {
				log.Error("Failed to encode response: %s", err.Error())
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
