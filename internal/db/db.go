package db

import (
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitDB(dsn string) (err error) {
	DB, err = sqlx.Connect("mysql", dsn)
	return
}
