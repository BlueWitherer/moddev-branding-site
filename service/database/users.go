package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"service/log"
	"service/utils"
)

func newUsers() *[]*utils.User {
	return new([]*utils.User)
}

// Current users cache
var currentUsers *[]*utils.User = nil
var currentUsersSince time.Time = time.Now()

func getUsers() *[]*utils.User {
	if currentUsers != nil {
		log.Debug("Returning cached imgs list")
		return currentUsers
	}

	currentUsersSince = time.Now()

	return newUsers()
}

func findUser(id int64) (*utils.User, bool) {
	if currentUsers != nil {
		for _, u := range *currentUsers {
			if u.ID == id {
				return u, true
			}
		}
	}

	return nil, false
}

func setUser(user *utils.User) *[]*utils.User {
	if currentUsers != nil {
		log.Debug("Caching user %s", user.ID)
		*currentUsers = append(*currentUsers, user)
	}

	return getUsers()
}

func deleteUser(id int64) *[]*utils.User {
	if currentUsers != nil {
		for i, u := range *currentUsers {
			if u.ID == id {
				*currentUsers = append((*currentUsers)[:i], (*currentUsers)[i+1:]...)
			}
		}
	}

	return getUsers()
}

func GetUser(id int64) (*utils.User, error) {
	if id == 0 {
		return nil, fmt.Errorf("empty user id")
	}

	if val, found := findUser(id); found {
		return val, nil
	}

	stmt, err := utils.PrepareStmt(dat, "SELECT * FROM users WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	user := new(utils.User)
	err = stmt.QueryRow(id).Scan(
		&user.ID,
		&user.Login,
		&user.AvatarURL,
		&user.IsAdmin,
		&user.IsStaff,
		&user.Verified,
		&user.Banned,
		&user.Created,
		&user.Updated,
	)
	if err != nil {
		return nil, err
	}

	currentUsers = setUser(user)

	return user, nil
}

func GetAllUsers() ([]*utils.User, error) {
	if time.Since(currentUsersSince) > 15*time.Minute {
		currentUsers = nil
	}

	if currentUsers != nil && len(*currentUsers) > 0 {
		log.Debug("Returning cached imgs list")
		return *getUsers(), nil
	}

	stmt, err := utils.PrepareStmt(dat, "SELECT * FROM users ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	users, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer users.Close()

	var out []*utils.User
	for users.Next() {
		u := new(utils.User)
		if err := users.Scan(
			&u.ID,
			&u.Login,
			&u.AvatarURL,
			&u.IsAdmin,
			&u.IsStaff,
			&u.Verified,
			&u.Banned,
			&u.Created,
			&u.Updated,
		); err != nil {
			return nil, err
		}

		currentUsers = setUser(u)

		out = append(out, u)
	}

	return out, users.Err()
}

// inserts a new user or updates login if it already exists.
func UpsertUser(id int64, login string, avatarUrl string) error {
	if id == 0 {
		return fmt.Errorf("empty user id")
	}

	stmt, err := utils.PrepareStmt(dat, "INSERT INTO users (id, login, avatar_url) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE login = VALUES (login), avatar_url = VALUES (avatar_url), updated_at = CURRENT_TIMESTAMP")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, login, avatarUrl)
	return err
}

func VerifyUser(id int64) (*utils.User, error) {
	stmt, err := utils.PrepareStmt(dat, "UPDATE users SET verified = TRUE WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return nil, err
	}

	return GetUser(id)
}

func StaffUser(id int64) (*utils.User, error) {
	stmt, err := utils.PrepareStmt(dat, "UPDATE users SET is_staff = TRUE WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return nil, err
	}

	return GetUser(id)
}

func BanUser(id int64) (*utils.User, error) {
	// delete all images associated with the user
	deleteImgsStmt, err := utils.PrepareStmt(dat, "SELECT * FROM images WHERE user_id = ?")
	if err != nil {
		return nil, err
	}
	defer deleteImgsStmt.Close()

	rows, err := deleteImgsStmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	imgs := make([]utils.Img, 0)
	for rows.Next() {
		var r utils.Img
		if err := rows.Scan(
			&r.ImgID,
			&r.UserID,
			&r.ImageURL,
			&r.Created,
			&r.Pending,
		); err != nil {
			return nil, err
		}

		imgs = append(imgs, r)
	}

	for _, a := range imgs {
		imgDir := filepath.Join("..", "cdn", fmt.Sprintf("%s-%d.webp", a.UserID, a.ImgID))
		err = os.Remove(imgDir)
		if err != nil {
			return nil, err
		}
	}

	user, err := GetUser(id)
	if err != nil {
		return nil, err
	}

	// ban the user
	stmt, err := utils.PrepareStmt(dat, "UPDATE users SET banned = TRUE WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return nil, err
	}

	currentUsers = deleteUser(id)

	return user, nil
}

func UnbanUser(id int64) (*utils.User, error) {
	// unban the user
	stmt, err := utils.PrepareStmt(dat, "UPDATE users SET banned = FALSE WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return nil, err
	}

	return GetUser(id)
}

func init() {
	users, err := GetAllUsers()
	if err != nil {
		log.Error("Failed to initialize users cache: %s", err.Error())
	} else {
		currentUsers = &users
		log.Info("Initialized users cache with %d users", len(users))
	}
}
