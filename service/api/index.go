package api

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gen2brain/webp"
	gwebp "github.com/gen2brain/webp"
	"github.com/patrickmn/go-cache"
	"github.com/samber/mo"

	"service/log"
	"service/utils"
)

var fixedUsernames = cache.New(12*time.Hour, 1*time.Hour)

func getGitUsername(repoUrl string) mo.Result[string] {
	u, err := url.Parse(repoUrl)
	if err != nil {
		return mo.Err[string](fmt.Errorf("invalid URL: %w", err))
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) <= 0 {
		return mo.Err[string](fmt.Errorf("invalid GitHub repo URL: %s", repoUrl))
	}

	return mo.Ok(parts[0])
}

func writeAsPNG(w io.Writer, src io.Reader, decode func(io.Reader) mo.Result[image.Image]) mo.Result[bool] {
	decodedRes := decode(src)
	if decodedRes.IsError() {
		return mo.Err[bool](fmt.Errorf("decode source image: %w", decodedRes.Error()))
	}
	img := decodedRes.MustGet()
	if err := png.Encode(w, img); err != nil {
		return mo.Err[bool](err)
	}
	return mo.Ok(true)
}

func writeAsWebp(w io.Writer, src io.Reader, decode func(io.Reader) mo.Result[image.Image]) mo.Result[bool] {
	decodedRes := decode(src)
	if decodedRes.IsError() {
		return mo.Err[bool](fmt.Errorf("decode source image: %w", decodedRes.Error()))
	}
	img := decodedRes.MustGet()
	if err := gwebp.Encode(w, img, gwebp.Options{Quality: 80}); err != nil {
		return mo.Err[bool](err)
	}
	return mo.Ok(true)
}

func decodeWebp(r io.Reader) mo.Result[image.Image] {
	img, err := webp.Decode(r)
	if err != nil {
		return mo.Err[image.Image](err)
	}
	return mo.Ok(img)
}

func decodePng(r io.Reader) mo.Result[image.Image] {
	img, err := png.Decode(r)
	if err != nil {
		return mo.Err[image.Image](err)
	}
	return mo.Ok(img)
}

func init() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Mod Developer Branding API service pinged")
		header := w.Header()

		utils.WriteHeaders(&header, http.MethodGet, false)
		utils.WriteWebRes(w, mo.Some("pong!"), http.StatusOK)
	})
}
