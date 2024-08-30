package tests

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/random"

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
// вернет статус 200 и сохранится в БД
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
	defer t.Cleanup(func() {
		_, err := conn.Exec(ctx, "delete from url where alias=$1", req.Alias)
		if err != nil {
			panic(err)
		}
	})
	require.NoError(t, err)
	var urlId int
	_ = conn.QueryRow(ctx, "select url_id from url where alias=$1", req.Alias).Scan(&urlId)
	assert.Greater(t, urlId, -1)
}
