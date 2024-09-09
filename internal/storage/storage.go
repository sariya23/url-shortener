package storage

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

var (
	ErrURLNotFound = errors.New("url not found")
	ErrAliasExists = errors.New("alias exists")
)

type Storage struct {
	connection *pgx.Conn
}

func MustNewConnection(ctx context.Context, storagePath string) (Storage, func(Storage), error) {
	const operationPlace = "storage.storage.MustNewConnection"
	cancel := func(s Storage) {
		err := s.connection.Close(ctx)
		if err != nil {
			panic(fmt.Sprintf("%v in %s", err, operationPlace))
		}
	}
	conn, err := pgx.Connect(ctx, storagePath)
	if err != nil {
		panic(fmt.Sprintf("%v in %s", err, operationPlace))
	}
	return Storage{connection: conn}, cancel, nil
}
