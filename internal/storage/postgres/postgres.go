package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"url-shortener/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Storage struct {
	connection *pgx.Conn
}

func MustNewConnection(ctx context.Context, storagePath string) (*Storage, func(s Storage), error) {
	const operationPlace = "storage.storage.MustNewConnection"
	cancel := func(s Storage) {
		err := s.connection.Close(ctx)
		if err != nil {
			log.Fatalf("Cannot close connection: %v (%s)", err, operationPlace)
		}
	}
	conn, err := pgx.Connect(ctx, storagePath)
	if err != nil {
		log.Fatalf("Cannot connect to db: %v (%s)", err, operationPlace)
	}
	err = conn.Ping(ctx)
	if err != nil {
		log.Fatalf("DB is not availavle: %v (%s)", err, operationPlace)
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
		return 0, fmt.Errorf("%s: %w", operationPlace, storage.ErrAliasExists)
	}

	if err != nil {
		return 0, fmt.Errorf("%s: %w", operationPlace, err)
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

func (s *Storage) DeleteURLByAlias(ctx context.Context, alias string) (int, error) {
	const operationPlace = "storage.postgres.DeleteURLByAlias"

	var deletedRows int

	query := `delete from url where alias=$1 returning url_id`
	err := s.connection.QueryRow(ctx, query, alias).Scan(&deletedRows)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return deletedRows, nil
}

func (s *Storage) DeleteURLByURL(ctx context.Context, url string) (int, error) {
	const operationPlace = "storage.postgres.DeleteURLByURL"

	var deletedRows int

	query := `delete from url where url=$1 returning url_id`
	err := s.connection.QueryRow(ctx, query, url).Scan(&deletedRows)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return deletedRows, nil
}

func (s *Storage) Truncate(ctx context.Context) error {
	const operationPlace = "storage.postgres.Truncate"
	query := `truncate url`
	_, err := s.connection.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %w", operationPlace, err)
	}
	return nil
}

func (s *Storage) GetURLIdByURL(ctx context.Context, URL string) (int, error) {
	const operationPlace = "storage.postgres.GetURLIdByURL"
	var urlId int

	query := `select url_id from url where url=$1`
	err := s.connection.QueryRow(ctx, query, URL).Scan(&urlId)

	if errors.Is(err, pgx.ErrNoRows) {
		return -1, storage.ErrURLNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return urlId, nil
}
