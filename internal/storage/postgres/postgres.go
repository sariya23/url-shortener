package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	Connection *pgx.Conn
}

func New(storagePath string) (*Storage, func(conn *pgx.Conn), error) {
	const operationPlace = "storage.postgres.New"
	cancel := func(conn *pgx.Conn) {
		err := conn.Close(context.Background())
		if err != nil {
			panic(err)
		}
	}
	conn, err := pgx.Connect(context.Background(), storagePath)

	if err != nil {
		return &Storage{Connection: conn}, cancel, fmt.Errorf("%s: %w", operationPlace, err)
	}

	_, err = conn.Query(context.Background(), `
	create table if not exists url (
		url_id bigint generated always as identity primary key,
		alias text not null unique,
		url text not null
	);
	`)

	if err != nil {
		return &Storage{Connection: conn}, cancel, fmt.Errorf("%s: %w", operationPlace, err)
	}

	return &Storage{Connection: conn}, cancel, nil
}
