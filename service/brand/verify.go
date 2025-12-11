package brand

import (
	"fmt"
	"net/http"
	"service/access"
	"service/database"
	"service/log"
	"strconv"
)

func init() {
	http.HandleFunc("/brand/verify", func(w http.ResponseWriter, r *http.Request) {
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

			u, err := database.GetUser(uid)
			if err != nil {
				log.Error("Failed to get ad owner: %s", err.Error())
				http.Error(w, "Failed to get ad owner", http.StatusInternalServerError)
				return
			}

			if !u.IsAdmin {
				log.Error("User of ID %s is not admin or staff", u.ID)
				http.Error(w, "User is not admin or staff", http.StatusUnauthorized)
				return
			}

			query := r.URL.Query()
			userStr := query.Get("user")

			userId, err := strconv.ParseUint(userStr, 10, 64)
			if err != nil {
				log.Error("Failed to get img ID: %s", err.Error())
				http.Error(w, "Failed to get img ID", http.StatusBadRequest)
				return
			}

			user, err := database.VerifyUser(userId)
			if err != nil {
				log.Error("Failed to verify user: %s", err.Error())
				http.Error(w, "Failed to verify user", http.StatusBadRequest)
				return
			}

			log.Info("Admin %s verified user %s", u.Login, user.Login)

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "User successfully verified")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
