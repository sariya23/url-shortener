package redirect_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogpretty/slogdiscard"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	caseName  string
	alias     string
	url       string
	respError string
	mockError error
}

func TestRedirectSuccess(t *testing.T) {
	cases := []testData{
		{
			caseName: "success redirect",
			alias:    "zxc",
			url:      "http://google.com",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.caseName, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)
			ctx := context.Background()
			urlGetterMock.On("GetURLByAlias", ctx, testCase.alias).Return(testCase.url, testCase.mockError).Once()
			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(ctx, slogdiscard.NewDiscardLogger(), urlGetterMock))
			server := httptest.NewServer(r)
			defer server.Close()

			redirect, err := api.GetRedirect(server.URL + "/" + testCase.alias)
			require.NoError(t, err)

			assert.Equal(t, testCase.url, redirect)
		})
	}
}

func TestRedirectFromAlias(t *testing.T) {
	cases := []testData{
		{
			caseName:  "No url on alias",
			alias:     "qwe",
			url:       "http://qwe.ru",
			respError: "no url on this alias (alias=qwe)",
			mockError: storage.ErrURLNotFound,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.caseName, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)
			ctx := context.Background()
			urlGetterMock.On("GetURLByAlias", ctx, testCase.alias).Return(testCase.url, testCase.mockError).Once()
			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(ctx, slogdiscard.NewDiscardLogger(), urlGetterMock))
			server := httptest.NewServer(r)
			defer server.Close()
			client := http.Client{}

			resp, err := client.Get(server.URL + "/" + testCase.alias)
			require.NoError(t, err)
			data, _ := io.ReadAll(resp.Body)
			var res response.Response
			err = json.Unmarshal(data, &res)
			require.NoError(t, err)

			assert.Equal(t, res.Error, testCase.respError)

		})
	}
}
