package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"service/database"
	"service/log"
	"service/utils"

	"github.com/patrickmn/go-cache"
)

func init() {
	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Mod Developer Branding API v1 service pinged")
		header := w.Header()

		utils.WriteHeaders(&header, http.MethodGet, false)

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong!")
	})

	http.HandleFunc("/api/v1/image", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Getting developer branding image...")
		header := w.Header()

		utils.WriteHeaders(&header, http.MethodGet, false)

		if r.Method == http.MethodGet {
			header.Set("Content-Type", "image/webp")

			query := r.URL.Query()

			dev := query.Get("dev")
			modId := query.Get("mod")

			fmtParam := query.Get("fmt")
			wantWebp := fmtParam == "webp"

			user, err := database.GetUserFromLogin(dev).Get()
			if err != nil {
				log.Warn("Failed to get user: %s", err.Error())

				if fixed, found := fixedUsernames.Get(dev); found {
					user, err = database.GetUserFromLogin(fixed.(string)).Get()
					if err != nil {
						log.Error("Failed to get user: %s", err.Error())
						http.Error(w, "Failed to get user", http.StatusNotFound)
						return
					}
				} else if modId != "" {
					mod, err := database.GetModCached(modId).Get()
					if err != nil {
						log.Error("Failed to get mod: %v", err)
						http.Error(w, "Failed to get mod", http.StatusNotFound)
						return
					}

					modDev, err := database.ResolveDevFromModID(mod.ID, dev).Get()
					if err != nil {
						log.Error("Failed to get mod developer: %v", err)
						http.Error(w, "Failed to get mod developer", http.StatusNotFound)
						return
					}

					user, err = database.GetUserFromLogin(modDev.Username).Get()
					if err != nil {
						log.Error("Failed to get user: %s", err.Error())
						http.Error(w, "Failed to get user", http.StatusNotFound)
						return
					}

					username, err := getGitUsername(mod.Links.Source).Get()
					if err != nil {
						log.Warn("Couldn't get GitHub username from repository URL %s", modDev.Username)
					} else if username != "" && dev != "" && username == dev {
						fixedUsernames.Set(username, modDev.Username, cache.DefaultExpiration)
					} else {
						log.Warn("Usernames %s and %s do not match or are empty", dev, modDev.Username)
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

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Error("Failed to read fallback image: %v", err)
						http.Error(w, "Failed to read image", http.StatusInternalServerError)
						return
					}

					if wantWebp {
						header.Set("Content-Type", "image/webp")
						w.WriteHeader(http.StatusOK)
						if _, err := writeAsWebp(w, bytes.NewReader(body), decodePng).Get(); err != nil {
							log.Error("Failed to convert fallback image to webp: %v", err)
						}
					} else {
						header.Set("Content-Type", "image/png")
						w.WriteHeader(http.StatusOK)
						if _, err := w.Write(body); err != nil {
							log.Error("Failed to stream fallback image: %v", err)
						}
					}

					return
				}
			}

			if user != nil {
				img, err := database.GetImageForUser(user.ID).Get()
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
				dstPath := filepath.Join("cdn", fileName)

				log.Info("Getting brand image %s for %s", dstPath, user.Login)

				f, err := os.Open(dstPath)
				if err != nil {
					log.Error("Failed to open image: %s", err.Error())
					http.Error(w, "Failed to open image", http.StatusNotFound)
					return
				}
				defer f.Close()

				if wantWebp {
					header.Set("Content-Type", "image/webp")
					w.WriteHeader(http.StatusOK)
					if _, err := io.Copy(w, f); err != nil {
						log.Error("Failed to stream image: %s", err.Error())
					}
					return
				}

				header.Set("Content-Type", "image/png")
				w.WriteHeader(http.StatusOK)
				if _, err := writeAsPNG(w, f, decodeWebp).Get(); err != nil {
					log.Error("Failed to convert image to png: %s", err.Error())
				}
				return
			} else {
				log.Error("Failed to process user")
				http.Error(w, "Failed to process user", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
