package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"service/log"
	"service/utils"

	"github.com/patrickmn/go-cache"
	"github.com/samber/mo"
)

func newImages() *[]*utils.Img {
	return new([]*utils.Img)
}

var currentImages *[]*utils.Img = nil
var currentImagesSince time.Time = time.Now()

func getImages() *[]*utils.Img {
	if currentImages != nil {
		log.Debug("Returning cached imgs list")
		return currentImages
	}

	currentImagesSince = time.Now()

	return newImages()
}

func findImage(id uint64) (*utils.Img, bool) {
	if currentImages != nil {
		for _, img := range *currentImages {
			if img.ID == id {
				return img, true
			}
		}
	}

	return nil, false
}

func findImageFromUser(id uint64) (*utils.Img, bool) {
	if currentImages != nil {
		for _, img := range *currentImages {
			if img.UserID == id {
				return img, true
			}
		}
	}

	return nil, false
}

func setImage(image *utils.Img) *[]*utils.Img {
	if currentImages != nil {
		log.Debug("Caching img %d", image.ID)
		for i, img := range *currentImages {
			if img.ID == image.ID {
				(*currentImages)[i] = img
				return getImages()
			}
		}

		*currentImages = append(*currentImages, image)
	}

	return getImages()
}

func deleteImage(id uint64) *[]*utils.Img {
	if currentImages != nil {
		for i, img := range *currentImages {
			if img.ID == id {
				*currentImages = append((*currentImages)[:i], (*currentImages)[i+1:]...)
			}
		}
	}

	return getImages()
}

func ApproveImage(id uint64) mo.Result[*utils.Img] {
	stmtRes := utils.PrepareStmt(dat, "UPDATE images SET pending = FALSE, created_at = NOW() WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.Img](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err := stmt.Exec(id)
	if err != nil {
		return mo.Err[*utils.Img](err)
	}

	if img, found := findImage(id); found {
		img.Pending = false
		currentImages = setImage(img)
	}

	imgRes := GetImage(id)
	if imgRes.IsError() {
		return mo.Err[*utils.Img](imgRes.Error())
	}
	img := imgRes.MustGet()
	return mo.Ok(img)
}

func CreateImage(userId uint64, url string) mo.Result[uint64] {
	if userId == 0 {
		return mo.Err[uint64](fmt.Errorf("missing img fields"))
	}

	stmtRes := utils.PrepareStmt(dat, "INSERT INTO images (user_id, image_url, pending) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE image_url = VALUES(image_url), pending = VALUES(pending), created_at = CURRENT_TIMESTAMP")
	if stmtRes.IsError() {
		return mo.Err[uint64](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	res, err := stmt.Exec(userId, url, true)
	if err != nil {
		return mo.Err[uint64](err)
	}

	if img, found := findImageFromUser(userId); found {
		img.ImageURL = url
		currentImages = setImage(img)
	}

	last, err := res.LastInsertId()
	if err != nil {
		return mo.Err[uint64](err)
	}
	return mo.Ok(uint64(last))
}

func ListAllImages() mo.Result[[]*utils.Img] {
	if time.Since(currentImagesSince) > 15*time.Minute {
		currentImages = nil
	}

	if currentImages != nil && len(*currentImages) > 0 {
		log.Debug("Returning cached imgs list")
		return mo.Ok(*getImages())
	}

	stmtRes := utils.PrepareStmt(dat, "SELECT * FROM images ORDER BY id DESC")
	if stmtRes.IsError() {
		return mo.Err[[]*utils.Img](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return mo.Err[[]*utils.Img](err)
	}
	defer rows.Close()

	var out []*utils.Img
	for rows.Next() {
		r := new(utils.Img)
		if err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.ImageURL,
			&r.Created,
			&r.Pending,
		); err != nil {
			return mo.Err[[]*utils.Img](err)
		}

		currentImages = setImage(r)

		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		return mo.Err[[]*utils.Img](err)
	}
	return mo.Ok(out)
}

func ListPendingImages() mo.Result[[]*utils.Img] {
	stmtRes := utils.PrepareStmt(dat, "SELECT * FROM images WHERE pending = TRUE ORDER BY id DESC")
	if stmtRes.IsError() {
		return mo.Err[[]*utils.Img](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return mo.Err[[]*utils.Img](err)
	}
	defer rows.Close()

	out := make([]*utils.Img, 0)
	for rows.Next() {
		r := new(utils.Img)
		if err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.ImageURL,
			&r.Created,
			&r.Pending,
		); err != nil {
			return mo.Err[[]*utils.Img](err)
		}

		currentImages = setImage(r)

		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		return mo.Err[[]*utils.Img](err)
	}
	return mo.Ok(out)
}

func FilterImagesByPending(rows []*utils.Img, showPending bool) mo.Result[[]*utils.Img] {
	out := make([]*utils.Img, 0)
	for _, r := range rows {
		if r.Pending == showPending {
			out = append(out, r)
		}
	}

	return mo.Ok(out)
}

func FilterImagesFromBannedUsers(rows []*utils.Img) mo.Result[[]*utils.Img] {
	out := make([]*utils.Img, 0)
	for _, r := range rows {
		userRes := GetUser(r.UserID)
		if userRes.IsError() {
			return mo.Err[[]*utils.Img](userRes.Error())
		}
		user := userRes.MustGet()

		if !user.Banned {
			out = append(out, r)
		}
	}

	return mo.Ok(out)
}

func FilterImagesByUser(rows []*utils.Img, userId uint64) mo.Result[[]*utils.Img] {
	out := make([]*utils.Img, 0)
	for _, r := range rows {
		if r.UserID == userId {
			out = append(out, r)
		}
	}

	return mo.Ok(out)
}

func GetImage(imgId uint64) mo.Result[*utils.Img] {
	if val, found := findImage(imgId); found {
		return mo.Ok(val)
	}

	stmtRes := utils.PrepareStmt(dat, "SELECT * FROM images WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.Img](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	row := stmt.QueryRow(imgId)
	if row != nil {
		r := new(utils.Img)
		if err := row.Scan(
			&r.ID,
			&r.UserID,
			&r.ImageURL,
			&r.Created,
			&r.Pending,
		); err != nil {
			if err == sql.ErrNoRows {
				return mo.Err[*utils.Img](err)
			}

			return mo.Err[*utils.Img](err)
		}

		currentImages = setImage(r)

		return mo.Ok(r)
	} else {
		return mo.Err[*utils.Img](fmt.Errorf("img not found"))
	}
}

func GetImageForUser(userId uint64) mo.Result[*utils.Img] {
	if val, found := findImageFromUser(userId); found {
		return mo.Ok(val)
	}

	stmtRes := utils.PrepareStmt(dat, "SELECT * FROM images WHERE user_id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.Img](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	row := stmt.QueryRow(userId)
	if row != nil {
		r := new(utils.Img)
		if err := row.Scan(
			&r.ID,
			&r.UserID,
			&r.ImageURL,
			&r.Created,
			&r.Pending,
		); err != nil {
			if err == sql.ErrNoRows {
				return mo.Err[*utils.Img](err)
			}

			return mo.Err[*utils.Img](err)
		}

		currentImages = setImage(r)

		return mo.Ok(r)
	} else {
		return mo.Err[*utils.Img](fmt.Errorf("img not found"))
	}
}

func GetImageOwnerId(imgId uint64) mo.Result[uint64] {
	if val, found := findImage(imgId); found {
		return mo.Ok(val.UserID)
	}

	var uid uint64

	stmtRes := utils.PrepareStmt(dat, "SELECT user_id FROM images WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[uint64](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	err := stmt.QueryRow(imgId).Scan(&uid)
	if err != nil {
		return mo.Err[uint64](err)
	}

	return mo.Ok(uid)
}

func DeleteImage(imgId uint64) mo.Result[*utils.Img] {
	imgRes := GetImage(imgId)
	if imgRes.IsError() {
		return mo.Err[*utils.Img](imgRes.Error())
	}
	img := imgRes.MustGet()

	stmtRes := utils.PrepareStmt(dat, "DELETE FROM images WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.Img](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err := stmt.Exec(imgId)
	if err != nil {
		return mo.Err[*utils.Img](err)
	}

	adDir := filepath.Join("cdn", fmt.Sprintf("%d.webp", img.UserID))
	err = os.Remove(adDir)
	if err != nil {
		return mo.Err[*utils.Img](err)
	}

	currentImages = deleteImage(imgId)

	return mo.Ok(img)
}

var ModCache = cache.New(24*time.Hour, 8*time.Hour)

func GetModCached(modID string) mo.Result[*utils.Mod] {
	if modID == "" {
		return mo.Err[*utils.Mod](fmt.Errorf("no mod id provided"))
	}

	if cached, found := ModCache.Get(modID); found {
		mod := cached.(utils.Mod)
		return mo.Ok(&mod)
	}

	apiURL := fmt.Sprintf("https://api.geode-sdk.org/v1/mods/%s", modID)
	resp, err := http.Get(apiURL)
	if err != nil {
		return mo.Err[*utils.Mod](fmt.Errorf("failed to fetch mod info: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return mo.Err[*utils.Mod](fmt.Errorf("mod API returned status %d", resp.StatusCode))
	}

	var modReq utils.ModRequest
	if err := json.NewDecoder(resp.Body).Decode(&modReq); err != nil {
		return mo.Err[*utils.Mod](fmt.Errorf("failed to decode mod API response: %w", err))
	}

	ModCache.Set(modID, modReq.Payload, cache.DefaultExpiration)

	return mo.Ok(&modReq.Payload)
}

func ResolveDevFromModID(modID string, dev string) mo.Result[*utils.ModDeveloper] {
	modRes := GetModCached(modID)
	if modRes.IsError() {
		return mo.Err[*utils.ModDeveloper](modRes.Error())
	}
	mod := modRes.MustGet()

	for _, devInfo := range mod.Developers {
		if devInfo.IsOwner {
			return mo.Ok(&devInfo)
		}
	}

	return mo.Err[*utils.ModDeveloper](fmt.Errorf("developer %s not found in mod %s", dev, modID))
}

func init() {
	imgsRes := ListAllImages()
	if imgsRes.IsError() {
		log.Error("Failed to initialize imgs cache: %s", imgsRes.Error())
	} else {
		imgs := imgsRes.MustGet()
		currentImages = &imgs
		log.Info("Initialized imgs cache with %d imgs", len(imgs))
	}
}
