package tests

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/random"
	errStorage "url-shortener/internal/storage"
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
	ctx := context.Background()
	err := godotenv.Load("../config/.env")
	require.NoError(t, err)

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
	ctx := context.Background()
	err := godotenv.Load("../config/.env")
	require.NoError(t, err)

	req := save.Request{
		URL:   gofakeit.URL(),
		Alias: "",
	}
	defer func() {
		storage, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
		require.NoError(t, err)
		defer cancel(*storage)
		_, err = storage.DeleteURLByURL(ctx, req.URL)
		if err != nil {
			panic(err)
		}
	}()
	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	e.POST("/url").WithJSON(req).
		WithBasicAuth("localuser", "password").
		Expect().
		JSON().
		Object().
		ContainsKey("status").ContainsValue(response.StatusOK)
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
	ctx := context.Background()
	err := godotenv.Load("../config/.env")
	require.NoError(t, err)
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

// TestRedirectSuccess проверяет, что
// происходит переадресация по URL, которому
// соответсвует алиас.
func TestRedirectSuccess(t *testing.T) {
	ctx := context.Background()
	err := godotenv.Load("../config/.env")
	require.NoError(t, err)
	storage, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer cancel(*storage)

	URL := "https://google.com"
	alias := "google"

	_, err = storage.SaveURL(ctx, URL, alias)
	require.NoError(t, err)
	defer func() {
		_, err := storage.DeleteURLByAlias(ctx, alias)
		if err != nil {
			panic(err)
		}
	}()

	u := url.URL{Scheme: "http", Host: host, Path: alias}
	redirectTo, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, URL, redirectTo)
}

// TestCannotRedirectBecauseAliasEmpty проверяет,
// что при передаче несуществующзего алиаса в запросе, сервер
// вернет ошибку.
func TestCannotRedirectUndefinedAlias(t *testing.T) {
	alias := "qwe"
	u := url.URL{Scheme: "http", Host: host, Path: alias}
	res, err := api.SendGet(u.String())
	require.NoError(t, err)
	assert.Equal(t, res.Error, redirect.ErrMsgRedirectNoAlias)
}

// TestDeleteSuccess проверяет, что при отправке
// запроса с алиасом, который есть в БД, произойдет удаление.
func TestDeleteSuccess(t *testing.T) {
	ctx := context.Background()
	err := godotenv.Load("../config/.env")
	require.NoError(t, err)
	storage, cancel, err := postgres.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer cancel(*storage)

	URL := "https://google.com"
	alias := "TestDeleteSuccess"

	_, err = storage.SaveURL(ctx, URL, alias)
	require.NoError(t, err)

	u := url.URL{Scheme: "http", Host: host, Path: "url"}
	e := httpexpect.Default(t, u.String())
	e.DELETE("/"+alias).WithBasicAuth("localuser", "password").
		Expect().
		JSON().
		Object().
		ContainsKey("status").ContainsValue("OK")

	id, err := storage.GetURLByAlias(ctx, alias)
	require.Error(t, err, errStorage.ErrURLNotFound)
	require.Equal(t, id, "")
}
