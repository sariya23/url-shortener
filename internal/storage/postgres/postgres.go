package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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
