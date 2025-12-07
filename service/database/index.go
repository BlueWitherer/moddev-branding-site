package database

import (
	"database/sql"

	"service/utils"
)

var dat *sql.DB

func init() {
	dat = utils.Db()
}
