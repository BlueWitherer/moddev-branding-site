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
				log.Error("Failed to get ad owner: %s", userRes.Error())
				http.Error(w, "Failed to get ad owner", http.StatusInternalServerError)
				return
			}
			u := userRes.MustGet()

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

			verifiedUserRes := database.VerifyUser(userId)
			if verifiedUserRes.IsError() {
				log.Error("Failed to verify user: %s", verifiedUserRes.Error())
				http.Error(w, "Failed to verify user", http.StatusBadRequest)
				return
			}
			user := verifiedUserRes.MustGet()

			log.Info("Admin %s verified user %s", u.Login, user.Login)

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "User successfully verified")
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
