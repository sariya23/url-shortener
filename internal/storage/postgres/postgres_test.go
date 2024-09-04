package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"url-shortener/internal/storage"
	"url-shortener/internal/storage/postgres"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../../../.env.local")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	exitVal := m.Run()
	os.Exit(exitVal)
}

// TestCreateTable проверяет, что
// создание таблицы проходит без ошибок.
func TestCreateTable(t *testing.T) {
	ctx := context.Background()
	storge, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(*storge)
	if err != nil {
		t.Errorf("cannot create table url: (%v)", err)
	}
}

// TestInsertURLInTable проверяет, что вставка
// записи в таблицу происходит без ошибок.
func TestInsertURLInTable(t *testing.T) {
	ctx := context.Background()
	storage, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(*storage)
	if err != nil {
		t.Errorf("cannot create table url: (%v)", err)
	}

	_, err = storage.SaveURL(ctx, "http://qwe.ru", "TestInsertURLinTable")
	if err != nil {
		t.Errorf("cannot save url: (%v)", err)
	}
}

// TestCannotSaveURLBecauseAliasAlreadyInTable проверяет,
// что попытка вставки такого же алиаса приведет к ошибке.
func TestCannotSaveURLBecauseAliasAlreadyInTable(t *testing.T) {
	ctx := context.Background()
	strg, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(*strg)
	if err != nil {
		t.Errorf("cannot create table url: (%v)", err)
	}

	_, err = strg.SaveURL(ctx, "http://qwe.ru", "TestCannotSaveURLBecauseURLAlreadyInTable")
	if err != nil {
		t.Errorf("cannot save url: (%v)", err)
	}

	_, err = strg.SaveURL(ctx, "http://qwe.ru", "TestCannotSaveURLBecauseURLAlreadyInTable")
	if !errors.Is(err, storage.ErrAliasExists) {
		t.Errorf("unexpected error %v, expected %v", err, storage.ErrAliasExists)
	}
}

// TestCanGetURLByAlias проверяет
// получение соответсвующего URL по алиасу.
func TestCanGetURLByAlias(t *testing.T) {
	ctx := context.Background()
	strg, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(*strg)
	if err != nil {
		t.Errorf("cannot create table url: (%v)", err)
	}
	alias, url := "TestCanGetURLByAlias", "http://qwe.ru"
	_, err = strg.SaveURL(ctx, url, alias)
	if err != nil {
		t.Errorf("cannot save url: (%v)", err)
	}

	urlFromTable, err := strg.GetURLByAlias(ctx, alias)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if urlFromTable != url {
		t.Errorf("URL is not equal. Got %s, expected %s", urlFromTable, url)
	}
}

// TestCannotGetURLBecauseItNotExists проверяет, что
// попытка получить URL по несуществующему алиасу приведет к ошибке.
func TestCannotGetURLBecauseItNotExists(t *testing.T) {
	ctx := context.Background()
	strg, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(*strg)
	if err != nil {
		t.Errorf("cannot create table url: (%v)", err)
	}

	_, err = strg.GetURLByAlias(ctx, "TestCannotGetURLBecauseItNotExists")
	if !errors.Is(err, storage.ErrURLNotFound) {
		t.Errorf("unexpected error. Expect %v, got %v", storage.ErrURLNotFound, err)
	}
}

// TestCanDeleteURLByAliasFromTable проверяет,
// что удаление строки по алиасу проходит успешно
// как при существующей записи, так и не существующего.
func TestCanDeleteURLByAliasFromTable(t *testing.T) {
	cases := []struct {
		caseName    string
		alias       string
		url         string
		deletedRows int
	}{
		{"Delete 1 row from table", "TestCanDeleteURLByAliasFromTable", "http://qwe.ru", 1},
		{"Delete 0 rows from table", "empty", "empty", 0},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			strg, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
			defer cancel(*strg)
			if err != nil {
				t.Errorf("cannot create table url: (%v)", err)
			}
			_, err = strg.SaveURL(ctx, tc.url, tc.alias)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			_, err = strg.DeleteURLByAlias(ctx, tc.alias)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCanDeleteURLByAliasFromTable проверяет,
// что удаление строки по URL проходит успешно
// как при существующей записи, так и не существующего.
func TestCanDeleteURLByURLFromTable(t *testing.T) {
	cases := []struct {
		caseName string
		alias    string
		url      string
	}{
		{"Delete 1 row from table", "TestCanDeleteURLByAliasFromTable", "http://qwe.ru"},
		{"Delete 0 rows from table", "empty", "empty"},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			strg, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
			defer cancel(*strg)
			if err != nil {
				t.Errorf("cannot create table url: (%v)", err)
			}
			_, err = strg.SaveURL(ctx, tc.url, tc.alias)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			_, err = strg.DeleteURLByURL(ctx, tc.url)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCanTruncateTable проверяет,
// что очистка таблицы происходит без ошибок.
func TestCanTruncateTable(t *testing.T) {
	ctx := context.Background()
	strg, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(*strg)
	if err != nil {
		t.Errorf("cannot create table url: (%v)", err)
	}

	err = strg.Truncate(ctx)
	if err != nil {
		t.Errorf("cannot truncate table. Get error: (%v)", err)
	}
}

// TestCanGetURLIdByURL проверяет,
// что получение url_id по URl происходит без ошибок.
func TestCanGetURLIdByURL(t *testing.T) {
	ctx := context.Background()
	strg, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(*strg)
	if err != nil {
		t.Errorf("cannot create table url: (%v)", err)
	}

	alias, url := "TestCanGetURLIdByURL", "http://qwe.ru"

	insertedId, err := strg.SaveURL(ctx, url, alias)
	if err != nil {
		t.Errorf("unexpected error: (%v)", err)
	}

	idFromTable, err := strg.GetURLIdByURL(ctx, url)
	if err != nil {
		t.Errorf("unexpected error: (%v)", err)
	}

	if idFromTable != insertedId {
		t.Errorf("expected %d, got %d", insertedId, idFromTable)
	}
}

// TestCannotGetURLIdByURLBecauseItNotExists, что
// при попытке получить запись по несуществующему URL проиходит
// ошибка.
func TestCannotGetURLIdByURLBecauseItNotExists(t *testing.T) {
	ctx := context.Background()
	strg, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	defer cancel(*strg)
	if err != nil {
		t.Errorf("cannot create table url: (%v)", err)
	}

	_, err = strg.GetURLIdByURL(ctx, "url")
	if !errors.Is(err, storage.ErrURLNotFound) {
		t.Errorf("unexpected error: (%v)", err)
	}

}
