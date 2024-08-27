package postgres

import (
	"context"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Storage struct {
	connection *pgx.Conn
}

func New(ctx context.Context, storagePath string) (*Storage, func(s Storage), error) {
	const operationPlace = "storage.postgres.New"
	cancel := func(s Storage) {
		err := s.connection.Close(ctx)
		if err != nil {
			panic(err)
		}
	}
	conn, err := pgx.Connect(ctx, storagePath)

	if err != nil {
		return &Storage{connection: conn}, cancel, fmt.Errorf("%s: %w", operationPlace, err)
	}

	_, err = conn.Exec(ctx, `
	create table if not exists url (
		url_id bigint generated always as identity primary key,
		alias text not null unique,
		url text not null
	);
	`)

	if err != nil {
		return &Storage{connection: conn}, cancel, fmt.Errorf("%s: %w", operationPlace, err)
	}

	_, err = conn.Exec(ctx, `create unique index if not exists url_idx on url(alias)`)

	if err != nil {
		return &Storage{connection: conn}, cancel, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return &Storage{connection: conn}, cancel, nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int, error) {
	const operationPlace = "storage.postgres.SaveURL"
	var insertedId int
	var pgErr *pgconn.PgError

	query := "insert into url(url, alias) values ($1, $2) returning url_id"
	err := s.connection.QueryRow(ctx, query, urlToSave, alias).Scan(&insertedId)

	if err != nil {
		if ok := errors.As(err, &pgErr); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return -1, fmt.Errorf("%s: %w", operationPlace, storage.ErrURLExists)
		}
		return -1, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return insertedId, nil
}
