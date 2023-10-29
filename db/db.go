package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB

func InitDB(url string) error {
	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(1)

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    email VARCHAR(256),
    profilePic VARCHAR(1000) NULL
	);`

	_, err = db.Exec(createTableSQL)

	Db = db

	if err != nil {
		return err
	}

	return nil
}
