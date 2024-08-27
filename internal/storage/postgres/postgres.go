package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// TODO: Подумать, правильно ли будет сделать это через UPSERT
func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int, error) {
	const operationPlace = "storage.postgres.SaveURL"
	var insertedId int
	var pgErr *pgconn.PgError

	query := "insert into url(url, alias) values ($1, $2) returning url_id"
	err := s.connection.QueryRow(ctx, query, urlToSave, alias).Scan(&insertedId)

	if ok := errors.As(err, &pgErr); ok && pgErr.Code == pgerrcode.UniqueViolation {
		return -1, fmt.Errorf("%s: %w", operationPlace, storage.ErrURLExists)
	}

	if err != nil {
		return -1, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return insertedId, nil
}

func (s *Storage) GetURLByAlias(ctx context.Context, alias string) (string, error) {
	const operationPlace = "storage.postgres.GetURLByAlias"
	var urlByAlias string

	query := `select url from url where alias=$1`
	err := s.connection.QueryRow(ctx, query, alias).Scan(&urlByAlias)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", operationPlace, err)
	}

	return urlByAlias, nil
}

func (s *Storage) DeleteURLByAlias(ctx context.Context, alias string) (string, error) {
	const operationPlace = "storage.postgres.GetURLByAlias"

	query := `delete from url where alias=$1`
	t, err := s.connection.Exec(ctx, query, alias)

	if err != nil {
		return "", fmt.Errorf("%s: %w", operationPlace, err)
	}

	deletedRows := string(strings.Split(t.String(), " ")[1])

	return deletedRows, nil
}
