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
	"github.com/samber/mo"
)

func init() {
	http.HandleFunc("/api/v2", func(w http.ResponseWriter, r *http.Request) {

		log.Debug("Mod Developer Branding API v2 service pinged")
		header := w.Header()

		utils.WriteHeaders(&header, http.MethodGet, false)
		utils.WriteWebRes(w, mo.Some("pong!"), http.StatusOK)
	})

	http.HandleFunc("/api/v2/image", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Getting developer branding image...")
		header := w.Header()
		utils.WriteHeaders(&header, http.MethodGet, false)

		if r.Method == http.MethodGet {
			query := r.URL.Query()

			dev := query.Get("dev")
			modId := query.Get("mod")

			fmtParam := query.Get("fmt")
			wantWebp := fmtParam == "webp"

			userRes := database.GetUserFromLogin(dev)
			var user *utils.User
			if userRes.IsError() {
				log.Warn("Failed to get user: %s", userRes.Error())

				if fixed, found := fixedUsernames.Get(dev); found {
					userRes = database.GetUserFromLogin(fixed.(string))
					if userRes.IsError() {
						log.Error("Failed to get user: %s", userRes.Error())
						utils.WriteWebErr(w, "Failed to get user", http.StatusNotFound)
						return
					}
					user = userRes.MustGet()
				} else if modId != "" {
					modRes := database.GetModCached(modId)
					if modRes.IsError() {
						log.Error("Failed to get mod: %v", modRes.Error())
						utils.WriteWebErr(w, "Failed to get mod", http.StatusNotFound)
						return
					}
					mod := modRes.MustGet()

					modDevRes := database.ResolveDevFromModID(mod.ID, dev)
					if modDevRes.IsError() {
						log.Error("Failed to get mod developer: %v", modDevRes.Error())
						utils.WriteWebErr(w, "Failed to get mod developer", http.StatusNotFound)
						return
					}
					modDev := modDevRes.MustGet()

					userRes = database.GetUserFromLogin(modDev.Username)
					if userRes.IsError() {
						log.Error("Failed to get user: %s", userRes.Error())
						utils.WriteWebErr(w, "Failed to get user", http.StatusNotFound)
						return
					}
					user = userRes.MustGet()

					usernameRes := getGitUsername(mod.Links.Source)
					if usernameRes.IsError() {
						log.Warn("Couldn't get GitHub username from repository URL %s", modDev.Username)
					} else {
						username := usernameRes.MustGet()
						if username != "" && dev != "" && username == dev {
							fixedUsernames.Set(username, modDev.Username, cache.DefaultExpiration)
						} else {
							log.Warn("Usernames %s and %s do not match or are empty", dev, modDev.Username)
						}
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
						utils.WriteWebErr(w, "Image not found", http.StatusNotFound)
						return
					}
					defer resp.Body.Close()

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Error("Failed to read fallback image: %v", err)
						utils.WriteWebErr(w, "Failed to read image", http.StatusInternalServerError)
						return
					}

					if wantWebp {
						var converted bytes.Buffer
						convertRes := writeAsWebp(&converted, bytes.NewReader(body), decodePng)
						if convertRes.IsError() {
							log.Error("Failed to convert fallback image to webp: %v", convertRes.Error())
							utils.WriteWebErr(w, "Failed to convert fallback image to webp", http.StatusInternalServerError)
							return
						}
						utils.WriteWebRes(w, mo.Some(converted.Bytes()), http.StatusOK)
					} else {
						utils.WriteWebRes(w, mo.Some(body), http.StatusOK)
					}

					return
				}
			} else {
				user = userRes.MustGet()
			}

			if userRes.IsOk() {
				imgRes := database.GetImageForUser(user.ID)
				if imgRes.IsError() {
					log.Error("Failed to get image info: %s", imgRes.Error())
					utils.WriteWebErr(w, "Failed to get image info", http.StatusInternalServerError)
					return
				}
				img := imgRes.MustGet()

				if img.Pending {
					log.Error("Image still pending review")
					utils.WriteWebErr(w, "Image still pending review", http.StatusForbidden)
					return
				}

				fileName := fmt.Sprintf("%d.webp", user.ID)
				dstPath := filepath.Join("cdn", fileName)

				log.Info("Getting brand image %s for %s", dstPath, user.Login)

				body, err := os.ReadFile(dstPath)
				if err != nil {
					log.Error("Failed to open image: %s", err.Error())
					utils.WriteWebErr(w, "Failed to open image", http.StatusNotFound)
					return
				}

				if wantWebp {
					utils.WriteWebRes(w, mo.Some(body), http.StatusOK)
					return
				}

				var converted bytes.Buffer
				conversionRes := writeAsPNG(&converted, bytes.NewReader(body), decodeWebp)
				if conversionRes.IsError() {
					log.Error("Failed to convert image to png: %s", conversionRes.Error())
					utils.WriteWebErr(w, "Failed to convert image to png", http.StatusInternalServerError)
					return
				}
				utils.WriteWebRes(w, mo.Some(converted.Bytes()), http.StatusOK)
				return
			} else {
				log.Error("Failed to process user")
				utils.WriteWebErr(w, "Failed to process user", http.StatusInternalServerError)
				return
			}
		} else {
			utils.WriteWebErr(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
