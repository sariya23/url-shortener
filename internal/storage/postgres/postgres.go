package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	db *pgx.Conn
}

func New(storagePath string) (*Storage, error) {
	const operationPlace = "storage.postgres.New"
	conn, err := pgx.Connect(context.Background(), storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", operationPlace, err)
	}

	res, err := conn.Query(context.Background(), "select 'hello world'")
	fmt.Println(res)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return &Storage{db: conn}, nil
}
