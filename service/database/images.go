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
)

func newImages() *[]*utils.Img {
	return new([]*utils.Img)
}

// Current imgs cache
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

func ApproveImage(id uint64) (*utils.Img, error) {
	stmt, err := utils.PrepareStmt(dat, "UPDATE images SET pending = FALSE, created_at = NOW() WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return nil, err
	}

	if img, found := findImage(id); found {
		img.Pending = false
		currentImages = setImage(img)
	}

	return GetImage(id)
}

// upserts a brand image row
func CreateImage(userId uint64, url string) (uint64, error) {
	if userId == 0 {
		return 0, fmt.Errorf("missing img fields")
	}

	// Create new img - allow multiple imgs per user per type
	stmt, err := utils.PrepareStmt(dat, "INSERT INTO images (user_id, image_url, pending) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE image_url = VALUES(image_url), pending = VALUES(pending), created_at = CURRENT_TIMESTAMP")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(userId, url, true)
	if err != nil {
		return 0, err
	}

	if img, found := findImageFromUser(userId); found {
		img.ImageURL = url
		currentImages = setImage(img)
	}

	last, err := res.LastInsertId()
	return uint64(last), err
}

// fetches all imgs for a given user
func ListAllImages() ([]*utils.Img, error) {
	if time.Since(currentImagesSince) > 15*time.Minute {
		currentImages = nil
	}

	if currentImages != nil && len(*currentImages) > 0 {
		log.Debug("Returning cached imgs list")
		return *getImages(), nil
	}

	stmt, err := utils.PrepareStmt(dat, "SELECT * FROM images ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
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
			return nil, err
		}

		currentImages = setImage(r)

		out = append(out, r)
	}

	return out, rows.Err()
}

func ListPendingImages() ([]*utils.Img, error) {
	stmt, err := utils.PrepareStmt(dat, "SELECT * FROM images WHERE pending = TRUE ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
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
			return nil, err
		}

		currentImages = setImage(r)

		out = append(out, r)
	}

	return out, rows.Err()
}

func FilterImagesByPending(rows []*utils.Img, showPending bool) ([]*utils.Img, error) {
	out := make([]*utils.Img, 0)
	for _, r := range rows {
		if r.Pending == showPending {
			out = append(out, r)
		}
	}

	return out, nil
}

func FilterImagesFromBannedUsers(rows []*utils.Img) ([]*utils.Img, error) {
	out := make([]*utils.Img, 0)
	for _, r := range rows {
		user, err := GetUser(r.UserID)
		if err != nil {
			return nil, err
		}

		if !user.Banned {
			out = append(out, r)
		}
	}

	return out, nil
}

func FilterImagesByUser(rows []*utils.Img, userId uint64) ([]*utils.Img, error) {
	out := make([]*utils.Img, 0)
	for _, r := range rows {
		if r.UserID == userId {
			out = append(out, r)
		}
	}

	return out, nil
}

func GetImage(imgId uint64) (*utils.Img, error) {
	if val, found := findImage(imgId); found {
		return val, nil
	}

	stmt, err := utils.PrepareStmt(dat, "SELECT * FROM images WHERE id = ?")
	if err != nil {
		return nil, err
	}
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
				return nil, err
			}

			return nil, err
		}

		currentImages = setImage(r)

		return r, nil
	} else {
		return nil, fmt.Errorf("img not found")
	}
}

func GetImageForUser(userId uint64) (*utils.Img, error) {
	if val, found := findImageFromUser(userId); found {
		return val, nil
	}

	stmt, err := utils.PrepareStmt(dat, "SELECT * FROM images WHERE user_id = ?")
	if err != nil {
		return nil, err
	}
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
				return nil, err
			}

			return nil, err
		}

		currentImages = setImage(r)

		return r, nil
	} else {
		return nil, fmt.Errorf("img not found")
	}
}

// returns the owning user_id for a brand image
func GetImageOwnerId(imgId uint64) (uint64, error) {
	if val, found := findImage(imgId); found {
		return val.UserID, nil
	}

	var uid uint64

	stmt, err := utils.PrepareStmt(dat, "SELECT user_id FROM images WHERE id = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(imgId).Scan(&uid)
	if err != nil {
		return 0, err
	}

	return uid, nil
}

func DeleteImage(imgId uint64) (*utils.Img, error) {
	img, err := GetImage(imgId)
	if err != nil {
		return img, err
	}

	stmt, err := utils.PrepareStmt(dat, "DELETE FROM images WHERE id = ?")
	if err != nil {
		return img, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(imgId)
	if err != nil {
		return img, err
	}

	adDir := filepath.Join("..", "cdn", fmt.Sprintf("%d.webp", img.UserID))
	err = os.Remove(adDir)
	if err != nil {
		return img, err
	}

	currentImages = deleteImage(imgId)

	return img, nil
}

var ModCache = cache.New(24*time.Hour, 1*time.Hour)

func GetModCached(modID string) (*utils.Mod, error) {
	if modID == "" {
		return nil, fmt.Errorf("no mod id provided")
	}

	if cached, found := ModCache.Get(modID); found {
		mod := cached.(utils.Mod)
		return &mod, nil
	}

	apiURL := fmt.Sprintf("https://api.geode-sdk.org/v1/mods/%s", modID)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mod info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mod API returned status %d", resp.StatusCode)
	}

	var modReq utils.ModRequest
	if err := json.NewDecoder(resp.Body).Decode(&modReq); err != nil {
		return nil, fmt.Errorf("failed to decode mod API response: %w", err)
	}

	ModCache.Set(modID, modReq.Payload, cache.DefaultExpiration)

	return &modReq.Payload, nil
}

func ResolveDevFromModID(modID string, dev string) (*utils.ModDeveloper, error) {
	mod, err := GetModCached(modID)
	if err != nil {
		return nil, err
	}

	for _, devInfo := range mod.Developers {
		if devInfo.IsOwner {
			return &devInfo, nil
		}
	}

	return nil, fmt.Errorf("developer %s not found in mod %s", dev, modID)
}

func init() {
	imgs, err := ListAllImages()
	if err != nil {
		log.Error("Failed to initialize imgs cache: %s", err.Error())
	} else {
		currentImages = &imgs
		log.Info("Initialized imgs cache with %d imgs", len(imgs))
	}
}
