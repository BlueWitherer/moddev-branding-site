package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"service/log"
	"service/utils"

	"github.com/samber/mo"
)

func newUsers() *[]*utils.User {
	return new([]*utils.User)
}

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

func findUser(id uint64) (*utils.User, bool) {
	if currentUsers != nil {
		for _, u := range *currentUsers {
			if u.ID == id {
				return u, true
			}
		}
	}

	return nil, false
}

func findUserByLogin(login string) (*utils.User, bool) {
	if currentUsers != nil {
		for _, u := range *currentUsers {
			if u.Login == login {
				return u, true
			}
		}
	}

	return nil, false
}

func setUser(user *utils.User) *[]*utils.User {
	if currentUsers != nil {
		log.Debug("Caching user %d", user.ID)
		*currentUsers = append(*currentUsers, user)
	}

	return getUsers()
}

func deleteUser(id uint64) *[]*utils.User {
	if currentUsers != nil {
		for i, u := range *currentUsers {
			if u.ID == id {
				*currentUsers = append((*currentUsers)[:i], (*currentUsers)[i+1:]...)
			}
		}
	}

	return getUsers()
}

func GetUser(id uint64) mo.Result[*utils.User] {
	if id == 0 {
		return mo.Err[*utils.User](fmt.Errorf("empty user id"))
	}

	if val, found := findUser(id); found {
		return mo.Ok(val)
	}

	stmtRes := utils.PrepareStmt(dat, "SELECT * FROM users WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.User](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	user := new(utils.User)
	err := stmt.QueryRow(id).Scan(
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
		return mo.Err[*utils.User](err)
	}

	currentUsers = setUser(user)

	return mo.Ok(user)
}

func GetUserFromLogin(login string) mo.Result[*utils.User] {
	if login == "" {
		return mo.Err[*utils.User](fmt.Errorf("empty user id"))
	}

	if val, found := findUserByLogin(login); found {
		return mo.Ok(val)
	}

	stmtRes := utils.PrepareStmt(dat, "SELECT * FROM users WHERE login = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.User](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	user := new(utils.User)
	err := stmt.QueryRow(login).Scan(
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
		return mo.Err[*utils.User](err)
	}

	currentUsers = setUser(user)

	return mo.Ok(user)
}

func GetAllUsers() mo.Result[[]*utils.User] {
	if time.Since(currentUsersSince) > 15*time.Minute {
		currentUsers = nil
	}

	if currentUsers != nil && len(*currentUsers) > 0 {
		log.Debug("Returning cached imgs list")
		return mo.Ok(*getUsers())
	}

	stmtRes := utils.PrepareStmt(dat, "SELECT * FROM users ORDER BY id DESC")
	if stmtRes.IsError() {
		return mo.Err[[]*utils.User](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	users, err := stmt.Query()
	if err != nil {
		return mo.Err[[]*utils.User](err)
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
			return mo.Err[[]*utils.User](err)
		}

		currentUsers = setUser(u)

		out = append(out, u)
	}

	if err := users.Err(); err != nil {
		return mo.Err[[]*utils.User](err)
	}
	return mo.Ok(out)
}

func UpsertUser(id uint64, login string, avatarUrl string) mo.Result[bool] {
	if id == 0 {
		return mo.Err[bool](fmt.Errorf("empty user id"))
	}

	stmtRes := utils.PrepareStmt(dat, "INSERT INTO users (id, login, avatar_url) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE login = VALUES (login), avatar_url = VALUES (avatar_url), updated_at = CURRENT_TIMESTAMP")
	if stmtRes.IsError() {
		return mo.Err[bool](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err := stmt.Exec(id, login, avatarUrl)
	if err != nil {
		return mo.Err[bool](err)
	}
	return mo.Ok(true)
}

func VerifyUser(id uint64) mo.Result[*utils.User] {
	stmtRes := utils.PrepareStmt(dat, "UPDATE users SET verified = TRUE WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.User](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err := stmt.Exec(id)
	if err != nil {
		return mo.Err[*utils.User](err)
	}

	if user, found := findUser(id); found {
		user.Verified = true
		currentUsers = setUser(user)
	}

	approveStmtRes := utils.PrepareStmt(dat, "UPDATE images SET pending = FALSE WHERE user_id = ?")
	if approveStmtRes.IsError() {
		return mo.Err[*utils.User](approveStmtRes.Error())
	}
	approveStmt := approveStmtRes.MustGet()
	defer approveStmt.Close()

	_, err = approveStmt.Exec(id)
	if err != nil {
		return mo.Err[*utils.User](err)
	}

	imgsRes := FilterImagesByUser(*getImages(), id)
	if imgsRes.IsError() {
		return mo.Err[*utils.User](imgsRes.Error())
	}
	imgs := imgsRes.MustGet()

	for _, img := range imgs {
		img.Pending = false
		currentImages = setImage(img)
	}

	userRes := GetUser(id)
	if userRes.IsError() {
		return mo.Err[*utils.User](userRes.Error())
	}
	user := userRes.MustGet()
	return mo.Ok(user)
}

func StaffUser(id uint64) mo.Result[*utils.User] {
	stmtRes := utils.PrepareStmt(dat, "UPDATE users SET is_staff = TRUE WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.User](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err := stmt.Exec(id)
	if err != nil {
		return mo.Err[*utils.User](err)
	}

	userRes := GetUser(id)
	if userRes.IsError() {
		return mo.Err[*utils.User](userRes.Error())
	}
	user := userRes.MustGet()
	return mo.Ok(user)
}

func BanUser(id uint64) mo.Result[*utils.User] {
	deleteImgsStmtRes := utils.PrepareStmt(dat, "SELECT * FROM images WHERE user_id = ?")
	if deleteImgsStmtRes.IsError() {
		return mo.Err[*utils.User](deleteImgsStmtRes.Error())
	}
	deleteImgsStmt := deleteImgsStmtRes.MustGet()
	defer deleteImgsStmt.Close()

	img := new(utils.Img)
	err := deleteImgsStmt.QueryRow(id).Scan(&img.ID, &img.UserID, &img.ImageURL, &img.Created, &img.Pending)
	if err != nil {
		return mo.Err[*utils.User](err)
	}

	imgDir := filepath.Join("cdn", fmt.Sprintf("%d.webp", img.UserID))
	err = os.Remove(imgDir)
	if err != nil {
		return mo.Err[*utils.User](err)
	}

	userRes := GetUser(id)
	if userRes.IsError() {
		return mo.Err[*utils.User](userRes.Error())
	}
	user := userRes.MustGet()

	stmtRes := utils.PrepareStmt(dat, "UPDATE users SET banned = TRUE WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.User](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return mo.Err[*utils.User](err)
	}

	currentUsers = deleteUser(id)

	return mo.Ok(user)
}

func UnbanUser(id uint64) mo.Result[*utils.User] {
	stmtRes := utils.PrepareStmt(dat, "UPDATE users SET banned = FALSE WHERE id = ?")
	if stmtRes.IsError() {
		return mo.Err[*utils.User](stmtRes.Error())
	}
	stmt := stmtRes.MustGet()
	defer stmt.Close()

	_, err := stmt.Exec(id)
	if err != nil {
		return mo.Err[*utils.User](err)
	}

	userRes := GetUser(id)
	if userRes.IsError() {
		return mo.Err[*utils.User](userRes.Error())
	}
	user := userRes.MustGet()
	return mo.Ok(user)
}

func init() {
	usersRes := GetAllUsers()
	if usersRes.IsError() {
		log.Error("Failed to initialize users cache: %s", usersRes.Error())
	} else {
		users := usersRes.MustGet()
		currentUsers = &users
		log.Info("Initialized users cache with %d users", len(users))
	}
}
