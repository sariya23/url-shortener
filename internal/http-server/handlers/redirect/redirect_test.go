package redirect_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/logger/handlers/slogpretty/slogdiscard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectFromAlias(t *testing.T) {
	cases := []struct {
		caseName  string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			caseName: "Success redirect. Alias and URL in db",
			alias:    "zxc",
			url:      "https://google.com",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.caseName, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)
			ctx := context.Background()
			if testCase.respError == "" || testCase.mockError != nil {
				urlGetterMock.On("GetURLByAlias", ctx, testCase.alias).
					Return(testCase.url, testCase.mockError).Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(ctx, slogdiscard.NewDiscardLogger(), urlGetterMock))

			server := httptest.NewServer(r)

			defer server.Close()

			redirectedToURL, err := api.GetRedirect(server.URL + "/" + testCase.alias)
			require.NoError(t, err)

			assert.Equal(t, testCase.url, redirectedToURL)
		})
	}
}
