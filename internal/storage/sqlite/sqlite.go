package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const operationPlace = "storage.sqilte.New"
	db, err := sql.Open("sqlite3", storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", operationPlace, err)
	}

	statement, err := db.Prepare(`
	create table if not exists url (
		url_id integer primary key,
		alias text not null unique,
		url text not null
	);
	create index if not exists idx_alias on url(alias);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", operationPlace, err)
	}

	_, err = statement.Exec()

	if err != nil {
		return nil, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return &Storage{db: db}, nil
}
