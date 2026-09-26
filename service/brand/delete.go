package brand

import (
	"fmt"
	"net/http"
	"strconv"

	"service/access"
	"service/database"
	"service/log"
)

func init() {
	http.HandleFunc("/brand/delete", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Attempting to delete img(s)...")
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", http.MethodDelete)
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodDelete {
			header.Set("Content-Type", "application/json")

			uidRes := access.GetSessionUserID(r)
			if uidRes.IsError() {
				log.Error("Unauthorized access to /brand/delete: %s", uidRes.Error())
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			uid := uidRes.MustGet()

			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				http.Error(w, "Missing img ID parameter", http.StatusBadRequest)
				return
			}

			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid img ID parameter", http.StatusBadRequest)
				return
			}

			ownerRes := database.GetImageOwnerId(id)
			if ownerRes.IsError() {
				log.Error("Failed to get image owner: %s", ownerRes.Error())
				http.Error(w, "Failed to get image owner", http.StatusInternalServerError)
				return
			}
			ownerid := ownerRes.MustGet()

			userRes := database.GetUser(uid)
			if userRes.IsError() {
				log.Error("Failed to get user: %s", userRes.Error())
				http.Error(w, "Failed to get user:", http.StatusInternalServerError)
				return
			}
			user := userRes.MustGet()

			if user.IsAdmin || user.IsStaff || ownerid == user.ID {
				imgRes := database.DeleteImage(id)
				if imgRes.IsError() {
					log.Error("Failed to delete image: %s", imgRes.Error())
					http.Error(w, "Failed to delete image", http.StatusInternalServerError)
					return
				}
				img := imgRes.MustGet()

				log.Info("Deleted image of ID %d", img.ID)

				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "Image deleted successfully")
			} else {
				log.Error("Unauthorized deletion attempt for img ID %d by user %s", id, uid)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
