package access

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"service/database"
	"service/log"
	"service/utils"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

var sessionCache = cache.New(2*time.Hour, 10*time.Minute)

func generateSessionID() string {
	return uuid.New().String() // uuid v4
}

func isSecure(r *http.Request) bool {
	if r.TLS != nil || os.Getenv("ENV") == "production" {
		return true
	}

	return false
}

func SetSession(w http.ResponseWriter, user *GitHubUser, secure bool) (string, error) {
	sessionId := generateSessionID()
	session := &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	}

	if secure {
		session.SameSite = http.SameSiteNoneMode
	} else {
		session.SameSite = http.SameSiteLaxMode
	}

	stmt, err := utils.PrepareStmt(utils.Db(), "INSERT INTO sessions (session_id, user_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE user_id = VALUES(user_id);")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sessionId, user.ID)
	if err != nil {
		return "", err
	}

	log.Debug("Setting session cookie...")
	http.SetCookie(w, session)

	sessionCache.Set(sessionId, user, cache.DefaultExpiration)

	return sessionId, nil
}

func GetSessionFromId(id string) (*GitHubUser, error) {
	var user GitHubUser
	if val, found := sessionCache.Get(id); found {
		user = val.(GitHubUser)
	}

	stmt, err := utils.PrepareStmt(utils.Db(), "SELECT user_id FROM sessions WHERE session_id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	updStmt, err := utils.PrepareStmt(utils.Db(), "UPDATE sessions SET last_seen = CURRENT_TIMESTAMP WHERE session_id = ?")
	if err != nil {
		return nil, err
	}
	defer updStmt.Close()

	_, err = updStmt.Exec(id)
	if err != nil {
		return nil, err
	}

	u, err := database.GetUser(user.ID)
	if err != nil {
		return nil, err
	}

	user.Login = u.Login
	user.AvatarURL = u.AvatarURL

	return &user, nil
}

func GetSessionUserID(r *http.Request) (int64, error) {
	c, err := r.Cookie("session_id")
	if err != nil {
		return 0, err
	}

	u, err := GetSessionFromId(c.Value)
	if err != nil || u == nil {
		if err == nil {
			err = fmt.Errorf("no user in session")
		}

		return 0, err
	}

	return u.ID, nil
}

func GetSession(r *http.Request) (*GitHubUser, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, err
	}

	user, err := GetSessionFromId(cookie.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func CleanupExpiredSessions() error {
	stmt, err := utils.PrepareStmt(utils.Db(), "DELETE FROM sessions WHERE last_seen < NOW() - INTERVAL 30 DAY")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec()
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	log.Info("Expired sessions cleaned: %d", rowsAffected)

	return nil
}

var sessionCancel context.CancelFunc

func StopSessionCleanup() {
	if sessionCancel != nil {
		sessionCancel()
	}
}

func init() {
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		redirectURL := "https://github.com/login/oauth/authorize?client_id=" +
			os.Getenv("GITHUB_CLIENT_ID") +
			"&redirect_uri=" + os.Getenv("GITHUB_REDIRECT_URI") +
			"&scope=read:user"

		http.Redirect(w, r, redirectURL, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}

		data := url.Values{}
		data.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
		data.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
		data.Set("code", code)
		data.Set("redirect_uri", os.Getenv("GITHUB_REDIRECT_URI"))

		req, _ := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var tokenResp Token
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			http.Error(w, "Failed to decode token", http.StatusInternalServerError)
			return
		}

		// Fetch user info
		req, _ = http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
		req.Header.Set("Authorization", tokenResp.TokenType+" "+tokenResp.AccessToken)

		resp, err = client.Do(req)
		if err != nil {
			http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		user := new(GitHubUser)
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			log.Error("Failed to decode user info: %s", err.Error())
			http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
			return
		}

		// Upsert into your DB
		if err := database.UpsertUser(
			user.ID,
			user.Login,
			user.AvatarURL,
		); err != nil {
			log.Error("Failed to upsert user: %s", err.Error())
			http.Error(w, "Failed to upsert user", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		_, err = SetSession(w, user, isSecure(r))
		if err != nil {
			log.Error("Failed to set session: %s", err.Error())
			http.Error(w, "Failed to set session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err == nil {
			sessionCache.Delete(cookie.Value)
			log.Info("User %s logged out", cookie.Value)
		}

		secure := false
		if r.TLS != nil || os.Getenv("ENV") == "production" {
			secure = true
		}

		clearCookie := &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // bye bye cookie
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteNoneMode,
		}

		http.SetCookie(w, clearCookie)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Logged out successfully")
	})

	http.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if c, err := r.Cookie("session_id"); err == nil {
				log.Debug("/session request cookie: %s", c.Value)
			} else {
				log.Debug("/session request no cookie: %s", err.Error())
			}

			user, err := GetSession(r)
			if err != nil {
				log.Error(err.Error())
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			header := w.Header()

			header.Set("Content-Type", "application/json")
			if jb, err := json.Marshal(user); err == nil {
				log.Debug("/session returning user: %s", string(jb))
			} else {
				log.Debug("/session returning user: (failed to marshal)")
			}

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(user); err != nil {
				log.Error("Failed to encode response: %s", err.Error())
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
