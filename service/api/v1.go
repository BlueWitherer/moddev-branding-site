package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"service/database"
	"service/log"

	"github.com/patrickmn/go-cache"
)

var fixedUsernames = cache.New(6*time.Hour, 1*time.Hour)

func getGitUsername(repoURL string) (string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid GitHub repo URL: %s", repoURL)
	}

	return parts[0], nil
}

func init() {
	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Mod Developer Branding API v1 service pinged")
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET")
		header.Set("Access-Control-Allow-Headers", "Content-Type")
		header.Set("Content-Type", "text/plain")

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong!")
	})

	http.HandleFunc("/api/v1/image", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Getting developer branding image...")
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET")
		header.Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodGet {
			header.Set("Content-Type", "image/webp")

			query := r.URL.Query()

			dev := query.Get("dev")
			modId := query.Get("mod")

			fmtParam := query.Get("fmt")

			user, err := database.GetUserFromLogin(dev)
			if err != nil {
				log.Warn("Failed to get user: %s", err.Error())

				if fixed, found := fixedUsernames.Get(dev); found {
					user, err = database.GetUserFromLogin(fixed.(string))
					if err != nil {
						log.Error("Failed to get user: %s", err.Error())
						http.Error(w, "Failed to get user", http.StatusNotFound)
						return
					}
				} else if modId != "" {
					mod, err := database.GetModCached(modId)
					if err != nil {
						log.Error("Failed to get mod: %v", err)
						http.Error(w, "Failed to get mod", http.StatusNotFound)
						return
					}

					modDev, err := database.ResolveDevFromModID(mod.ID, dev)
					if err != nil {
						log.Error("Failed to get mod developer: %v", err)
						http.Error(w, "Failed to get mod developer", http.StatusNotFound)
						return
					}

					user, err = database.GetUserFromLogin(modDev.Username)
					if err != nil {
						log.Error("Failed to get user: %s", err.Error())
						http.Error(w, "Failed to get user", http.StatusNotFound)
						return
					}

					username, err := getGitUsername(mod.Links.Source)
					if err != nil {
						log.Warn("Couldn't get GitHub username from repository URL %s", modDev.Username)
					} else if username == dev {
						fixedUsernames.Set(username, modDev.Username, cache.DefaultExpiration)
					} else {
						log.Warn("Usernames %s and %s do not match", dev, modDev.Username)
					}
				} else {
					devLower := strings.ToLower(dev)
					githubURL := fmt.Sprintf(
						"https://raw.githubusercontent.com/Alphalaneous/ModDevBranding-Images/refs/heads/main/Images/%s.png",
						devLower,
					)

					resp, err := http.Get(githubURL)
					if err != nil || resp.StatusCode != http.StatusOK {
						log.Error("Image not found: %v", err)
						http.Error(w, "Image not found", http.StatusNotFound)
						return
					}
					defer resp.Body.Close()

					if fmtParam == "webp" {
						header.Set("Content-Type", "image/webp")
					} else {
						header.Set("Content-Type", "image/png")
					}

					w.WriteHeader(http.StatusOK)
					if _, err := io.Copy(w, resp.Body); err != nil {
						log.Error("Failed to stream fallback image: %v", err)
						http.Error(w, "Failed to stream image", http.StatusInternalServerError)
						return
					}
				}
			}

			img, err := database.GetImageForUser(user.ID)
			if err != nil {
				log.Error("Failed to get image info: %s", err.Error())
				http.Error(w, "Failed to get image info", http.StatusInternalServerError)
				return
			}

			if img.Pending {
				log.Error("Image still pending review")
				http.Error(w, "Image still pending review", http.StatusForbidden)
				return
			}

			fileName := fmt.Sprintf("%d.webp", user.ID)
			dstPath := filepath.Join("..", "cdn", fileName)

			log.Info("Getting brand image %s for %s", dstPath, user.Login)

			f, err := os.Open(dstPath)
			if err != nil {
				log.Error("Failed to open image: %s", err.Error())
				http.Error(w, "Failed to open image", http.StatusNotFound)
				return
			}
			defer f.Close()

			if fmtParam == "webp" {
				header.Set("Content-Type", "image/webp")

				w.WriteHeader(http.StatusOK)
				if _, err := io.Copy(w, f); err != nil {
					log.Error("Failed to stream image: %s", err.Error())
					http.Error(w, "Failed to stream image", http.StatusInternalServerError)
					return
				}
			} else {
				header.Set("Content-Type", "image/png")

				w.WriteHeader(http.StatusOK)
				if _, err := io.Copy(w, f); err != nil {
					log.Error("Failed to stream image: %s", err.Error())
					http.Error(w, "Failed to stream image", http.StatusInternalServerError)
					return
				}
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
