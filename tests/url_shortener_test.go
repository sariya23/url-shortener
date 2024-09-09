package tests

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/random"
	errStorage "url-shortener/internal/storage"
	"url-shortener/internal/storage/postgres"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	cfg = map[string]string{
		"username": "localuser",
		"password": "password",
	}
)

const (
	host = "127.0.0.1:8082"
)

func TestMain(m *testing.M) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	err := godotenv.Load("../.env.local")
	if err != nil {
		logger.Fatalf("cannot load env: %v", err)
	}
	ctx := context.Background()
	dbPath := os.Getenv("DATABASE_URL")
	storage, cancel, err := postgres.MustNewConnection(ctx, dbPath)
	defer cancel(*storage)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("TEST DATABASE CREATED SUCCESS")
	exitVal := m.Run()
	logger.Println("TESTS COMPLETED")
	err = storage.Truncate(ctx)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("TEST DATABASE TRUNCATE")
	os.Exit(exitVal)
}

// TestSaveURLSuccess проверяет, что
// запрос на сохранение URL при переданном
// алиасе и при успешной аутентификация
// вернет статус 200 и сохранится в БД.
func TestSaveURLSuccess(t *testing.T) {
	ctx := context.Background()

	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	req := save.Request{
		URL:   gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}
	e.POST("/url").WithJSON(req).WithBasicAuth(cfg["username"], cfg["password"]).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ContainsKey("alias").ContainsValue(req.Alias)

	conn, cancel, err := postgres.MustNewConnection(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer cancel(*conn)
	URL, err := conn.GetURLByAlias(ctx, req.Alias)
	require.NoError(t, err)
	assert.Equal(t, req.URL, URL)
}

// TestSaveURLWithAliasByAutoGenerate проверяет,
// что если алиас не передан, то он генерируется автоматически
// и URL сохраняется с ним.
func TestSaveURLWithAliasByAutoGenerate(t *testing.T) {
	ctx := context.Background()

	req := save.Request{
		URL:   gofakeit.URL(),
		Alias: "",
	}
	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	e.POST("/url").WithJSON(req).
		WithBasicAuth(cfg["username"], cfg["password"]).
		Expect().
		JSON().
		Object().
		ContainsKey("status").ContainsValue(response.StatusOK)
	conn, cancel, err := postgres.MustNewConnection(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer cancel(*conn)
	urlId, err := conn.GetURLIdByURL(ctx, req.URL)
	require.NoError(t, err)
	assert.NotEqual(t, urlId, -1)
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
		URL:   "URL_TestCannotSaveInvalidURL",
		Alias: random.NewRandomString(10),
	}
	e.POST("/url").WithJSON(req).
		WithBasicAuth(cfg["username"], cfg["password"]).
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
	storage, cancel, err := postgres.MustNewConnection(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer cancel(*storage)
	alias := "ALIAS_TestCannotSaveTwoEqaulAliases"
	_, err = storage.SaveURL(ctx, "http://qwe.ru", alias)
	require.NoError(t, err)

	u := url.URL{Scheme: "http", Host: host}
	e := httpexpect.Default(t, u.String())
	req := save.Request{
		URL:   "http://TestCannotSaveTwoEqaulAliases.io",
		Alias: alias,
	}
	e.POST("/url").WithJSON(req).WithBasicAuth(cfg["username"], cfg["password"]).
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
	storage, cancel, err := postgres.MustNewConnection(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer cancel(*storage)

	URL := "https://google.com"
	alias := "ALIAS_TestRedirectSuccess"

	_, err = storage.SaveURL(ctx, URL, alias)
	require.NoError(t, err)

	u := url.URL{Scheme: "http", Host: host, Path: alias}
	redirectTo, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, URL, redirectTo)
}

// TestCannotRedirectBecauseAliasEmpty проверяет,
// что при передаче несуществующзего алиаса в запросе, сервер
// вернет ошибку.
func TestCannotRedirectUndefinedAlias(t *testing.T) {
	alias := "ALIAS_TestCannotRedirectUndefinedAlias"
	u := url.URL{Scheme: "http", Host: host, Path: alias}
	res, err := api.SendGet(u.String())
	require.NoError(t, err)
	assert.Equal(t, res.Error, redirect.ErrMsgRedirectNoAlias)
}

// TestDeleteSuccess проверяет, что при отправке
// запроса с алиасом, который есть в БД, произойдет удаление.
func TestDeleteSuccess(t *testing.T) {
	ctx := context.Background()
	storage, cancel, err := postgres.MustNewConnection(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer cancel(*storage)

	URL := "https://google.com"
	alias := "ALIAS_TestDeleteSuccess"

	_, err = storage.SaveURL(ctx, URL, alias)
	require.NoError(t, err)

	u := url.URL{Scheme: "http", Host: host, Path: "url"}
	e := httpexpect.Default(t, u.String())
	e.DELETE("/"+alias).WithBasicAuth(cfg["username"], cfg["password"]).
		Expect().
		JSON().
		Object().
		ContainsKey("status").ContainsValue("OK")

	id, err := storage.GetURLByAlias(ctx, alias)
	require.Error(t, err, errStorage.ErrURLNotFound)
	require.Equal(t, id, "")
}

// TestCannotDeleteBecauseNoRow проверяет, что
// сервер вернет ошибку, если в БД нет записи с этим
// алиасом.
func TestCannotDeleteBecauseNoRow(t *testing.T) {
	alias := "TestCannotDeleteBecauseNoRow"

	u := url.URL{Scheme: "http", Host: host, Path: "url"}
	e := httpexpect.Default(t, u.String())
	e.DELETE("/"+alias).WithBasicAuth(cfg["username"], cfg["password"]).
		Expect().
		JSON().
		Object().
		ContainsKey("error").ContainsValue(delete.ErrNothingToDelete)
}

// TestCannotDeleteRowWithOutAuth проверяет,
// что удаление невозможно без авторизации.
func TestCannotDeleteRowWithOutAuth(t *testing.T) {
	u := url.URL{Scheme: "http", Host: host, Path: "url"}
	e := httpexpect.Default(t, u.String())
	alias := "TestCannotDeleteRowWithOutAuth"
	e.DELETE("/" + alias).Expect().Status(http.StatusUnauthorized)
}
