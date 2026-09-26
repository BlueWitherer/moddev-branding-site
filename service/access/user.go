package access

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"service/database"
	"service/log"
	"service/utils"

	"github.com/patrickmn/go-cache"
	"github.com/samber/mo"
)

type GitHubUser struct {
	ID        uint64    `json:"id"`
	Login     string    `json:"login"`
	AvatarURL string    `json:"avatar_url"`
	IsAdmin   bool      `json:"is_admin"`
	IsStaff   bool      `json:"is_staff"`
	Verified  bool      `json:"verified"`
	Banned    bool      `json:"banned"`
	Created   time.Time `json:"created_at"`
	Updated   time.Time `json:"updated_at"`
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

var sessionCache = cache.New(2*time.Hour, 10*time.Minute)

type sessionIDs struct {
	id   string
	hash string
}

func generateSessionID() mo.Result[sessionIDs] {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return mo.Err[sessionIDs](err)
	}

	raw := base64.RawURLEncoding.EncodeToString(b)

	h := sha256.Sum256([]byte(raw))
	hash := base64.RawURLEncoding.EncodeToString(h[:])

	return mo.Ok(sessionIDs{id: raw, hash: hash})
}

func hashSessionID(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func isSecure(r *http.Request) bool {
	if r.TLS != nil || os.Getenv("ENV") == "production" {
		return true
	}

	return false
}

func SetSession(w http.ResponseWriter, user *GitHubUser, secure bool) mo.Result[string] {
	idsRes := generateSessionID()
	if idsRes.IsError() {
		return mo.Err[string](idsRes.Error())
	}
	ids := idsRes.MustGet()

	session := &http.Cookie{
		Name:     "session_id",
		Value:    ids.id,
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

	stmtRes := utils.PrepareStmt(utils.Db(), "INSERT INTO sessions (session_id, user_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE user_id = VALUES(user_id);")
	if stmtRes.IsError() {
		return mo.Err[string](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err := stmt.Exec(ids.hash, user.ID)
	if err != nil {
		return mo.Err[string](err)
	}

	log.Debug("Setting session cookie...")
	http.SetCookie(w, session)

	sessionCache.Set(ids.hash, user, cache.DefaultExpiration)

	return mo.Ok(ids.hash)
}

func GetSessionFromId(id string) mo.Result[*GitHubUser] {
	sessionId := hashSessionID(id)

	if val, found := sessionCache.Get(sessionId); found {
		if user, ok := val.(*GitHubUser); ok {
			return mo.Ok(user)
		}
	}
	var user GitHubUser

	stmtRes := utils.PrepareStmt(utils.Db(), "SELECT user_id FROM sessions WHERE session_id = ?")
	if stmtRes.IsError() {
		return mo.Err[*GitHubUser](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	err := stmt.QueryRow(sessionId).Scan(&user.ID)
	if err != nil {
		return mo.Err[*GitHubUser](err)
	}

	updStmtRes := utils.PrepareStmt(utils.Db(), "UPDATE sessions SET last_seen = CURRENT_TIMESTAMP WHERE session_id = ?")
	if updStmtRes.IsError() {
		return mo.Err[*GitHubUser](updStmtRes.Error())
	}
	updStmt := updStmtRes.MustGet()
	defer updStmt.Close()

	_, err = updStmt.Exec(sessionId)
	if err != nil {
		return mo.Err[*GitHubUser](err)
	}

	userRes := database.GetUser(user.ID)
	if userRes.IsError() {
		return mo.Err[*GitHubUser](userRes.Error())
	}
	u := userRes.MustGet()

	user.Login = u.Login
	user.AvatarURL = u.AvatarURL
	user.IsAdmin = u.IsAdmin
	user.IsStaff = u.IsStaff
	user.Verified = u.Verified
	user.Banned = u.Banned
	user.Created = u.Created
	user.Updated = u.Updated
	return mo.Ok(&user)
}

func GetSessionUserID(r *http.Request) mo.Result[uint64] {
	c, err := r.Cookie("session_id")
	if err != nil {
		return mo.Err[uint64](err)
	}

	userRes := GetSessionFromId(c.Value)
	if userRes.IsError() {
		return mo.Err[uint64](userRes.Error())
	}
	u := userRes.MustGet()
	if u == nil {
		return mo.Err[uint64](fmt.Errorf("no user in session"))
	}

	return mo.Ok(u.ID)
}

func GetSession(r *http.Request) mo.Result[*GitHubUser] {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return mo.Err[*GitHubUser](err)
	}

	userRes := GetSessionFromId(cookie.Value)
	if userRes.IsError() {
		return mo.Err[*GitHubUser](userRes.Error())
	}
	user := userRes.MustGet()

	return mo.Ok(user)
}

type sessionDeleteError struct {
	status int
	err    error
}

func (e sessionDeleteError) Error() string { return e.err.Error() }

func DeleteSession(r *http.Request) mo.Result[int] {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return mo.Err[int](sessionDeleteError{status: http.StatusUnauthorized, err: err})
	}

	stmtRes := utils.PrepareStmt(utils.Db(), "DELETE FROM sessions WHERE session_id = ?")
	if stmtRes.IsError() {
		return mo.Err[int](sessionDeleteError{status: http.StatusInternalServerError, err: stmtRes.Error()})
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err = stmt.Exec(hashSessionID(cookie.Value))
	if err != nil {
		return mo.Err[int](sessionDeleteError{status: http.StatusInternalServerError, err: err})
	}

	sessionCache.Delete(hashSessionID((cookie.Value)))

	return mo.Ok(http.StatusOK)
}

func CleanupExpiredSessions() mo.Result[bool] {
	stmtRes := utils.PrepareStmt(utils.Db(), "DELETE FROM sessions WHERE last_seen < NOW() - INTERVAL 30 DAY")
	if stmtRes.IsError() {
		return mo.Err[bool](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	res, err := stmt.Exec()
	if err != nil {
		return mo.Err[bool](err)
	}

	rowsAffected, _ := res.RowsAffected()
	log.Info("Expired sessions cleaned: %d", rowsAffected)

	return mo.Ok(true)
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

		upsertRes := database.UpsertUser(
			user.ID,
			user.Login,
			user.AvatarURL,
		)
		if upsertRes.IsError() {
			log.Error("Failed to upsert user: %s", upsertRes.Error())
			http.Error(w, "Failed to upsert user", http.StatusInternalServerError)
			return
		}

		sessionRes := SetSession(w, user, isSecure(r))
		if sessionRes.IsError() {
			log.Error("Failed to set session: %s", sessionRes.Error())
			http.Error(w, "Failed to set session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		sessionRes := DeleteSession(r)
		if sessionRes.IsError() {
			log.Error("Failed to log out: %s")
			status := http.StatusInternalServerError
			var deleteErr sessionDeleteError
			if errors.As(sessionRes.Error(), &deleteErr) {
				status = deleteErr.status
			}
			http.Error(w, "Failed to log out", status)
			return
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

			userRes := GetSession(r)
			if userRes.IsError() {
				log.Error(userRes.Error().Error())
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			user := userRes.MustGet()

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
