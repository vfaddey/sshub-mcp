package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func Open(dsnPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsnPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
