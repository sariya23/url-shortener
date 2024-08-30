package tests

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage/postgres"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8082"
)

// TestSaveURLSuccess проверяет, что
// запрос на сохранение URL при переданном
// алиасе и при успешной аутентификация
// вернет статус 200 и сохранится в БД.
func TestSaveURLSuccess(t *testing.T) {
	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	req := save.Request{
		URL:   gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}
	e.POST("/url").WithJSON(req).WithBasicAuth("localuser", "password").
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ContainsKey("alias").ContainsValue(req.Alias)

	err := godotenv.Load("../config/.env")
	require.NoError(t, err)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	defer func() {
		_, err := conn.Exec(ctx, "delete from url where url=$1", req.URL)
		if err != nil {
			panic(err)
		}
	}()
	require.NoError(t, err)
	var urlId int
	_ = conn.QueryRow(ctx, "select url_id from url where alias=$1", req.Alias).Scan(&urlId)
	assert.Greater(t, urlId, -1)
}

// TestSaveURLWithAliasByAutoGenerate проверяет,
// что если алиас не передан, то он генерируется автоматически
// и URL сохраняется с ним.
func TestSaveURLWithAliasByAutoGenerate(t *testing.T) {
	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	req := save.Request{
		URL:   gofakeit.URL(),
		Alias: "",
	}
	e.POST("/url").WithJSON(req).
		WithBasicAuth("localuser", "password").
		Expect().
		JSON().
		Object().
		ContainsKey("status").ContainsValue(response.StatusOK)
	err := godotenv.Load("../config/.env")
	require.NoError(t, err)
	ctx := context.Background()
	defer func() {
		storage, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
		require.NoError(t, err)
		defer cancel(*storage)
		_, err = storage.DeleteURLByURL(ctx, req.URL)
		if err != nil {
			panic(err)
		}
	}()
	require.NoError(t, err)
}

// TestCannotSaveURLWithoutAuth проверяет, что
// неавторизованный пользователь не сможет
// отправить запрос на сохранение.
func TestCannotSaveURLWithoutAuth(t *testing.T) {
	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	req := save.Request{
		URL:   gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}
	e.POST("/url").WithJSON(req).
		Expect().Status(http.StatusUnauthorized)
}

// TestCannotSaveInvalidURL проверяет, что
// если отправить невалидный URL, то сервер
// ответит ошибкой.
func TestCannotSaveInvalidURL(t *testing.T) {
	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	req := save.Request{
		URL:   "aboba",
		Alias: random.NewRandomString(10),
	}
	e.POST("/url").WithJSON(req).
		WithBasicAuth("localuser", "password").
		Expect().
		JSON().
		Object().
		ContainsKey("error").
		ContainsValue("field is not a valid URL. Field: URL")
}

// TestCannotSaveTwoEqaulURLs проверяет, что при
// попытке сохранения алиаса, который уже есть в БД
// сервер вернет ошибку.
func TestCannotSaveTwoEqaulAliases(t *testing.T) {
	err := godotenv.Load("../config/.env")
	require.NoError(t, err)
	ctx := context.Background()
	storage, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer cancel(*storage)
	alias := "abobaTEST"
	_, err = storage.SaveURL(ctx, "http://qwe.ru", alias)
	require.NoError(t, err)
	defer func() {
		_, err := storage.DeleteURLByAlias(ctx, alias)
		if err != nil {
			panic(err)

		}
	}()

	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	req := save.Request{
		URL:   "http://urlfortestABOBA.io",
		Alias: alias,
	}
	e.POST("/url").WithJSON(req).WithBasicAuth("localuser", "password").
		Expect().
		JSON().
		Object().
		ContainsKey("error").
		ContainsValue("alias already exists")

}
