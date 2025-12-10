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
		header.Set("Access-Control-Allow-Methods", "DELETE")
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodDelete {
			header.Set("Content-Type", "application/json")

			uid, err := access.GetSessionUserID(r)
			if err != nil {
				log.Error("Unauthorized access to /brand/delete: %s", err.Error())
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

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

			ownerid, err := database.GetImageOwnerId(id)
			if err != nil {
				log.Error("Failed to get image owner: %s", err.Error())
				http.Error(w, "Failed to get image owner", http.StatusInternalServerError)
				return
			}

			user, err := database.GetUser(uid)
			if err != nil {
				log.Error("Failed to get user: %s", err.Error())
				http.Error(w, "Failed to get user:", http.StatusInternalServerError)
				return
			}

			if user.IsAdmin || user.IsStaff || ownerid == user.ID {
				img, err := database.DeleteImage(id)
				if err != nil {
					log.Error("Failed to delete image: %s", err.Error())
					http.Error(w, "Failed to delete image", http.StatusInternalServerError)
					return
				}

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
