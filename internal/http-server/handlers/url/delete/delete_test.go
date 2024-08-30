package delete_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogpretty/slogdiscard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteQuerySuccess(t *testing.T) {
	cases := []struct {
		caseName  string
		alias     string
		respError string
		mockError error
	}{
		{
			caseName: "Success delete row",
			alias:    "qwe",
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			urlDeleterMock := mocks.NewURLDeleter(t)
			urlDeleterMock.On("DeleteURLByAlias", ctx, tc.alias).Return(1, tc.mockError)
			handler := delete.New(ctx, slogdiscard.NewDiscardLogger(), urlDeleterMock)
			r := chi.NewRouter()
			r.Delete("/{alias}", handler)
			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/"+tc.alias, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			defer resp.Body.Close()

			var respBody delete.Response

			err = json.Unmarshal(body, &respBody)
			require.NoError(t, err)

			assert.Equal(t, response.StatusOK, respBody.Status)
			assert.Equal(t, 1, respBody.DeletedId)
			assert.Equal(t, tc.alias, respBody.Alias)
			assert.Equal(t, "", respBody.Error)
		})
	}
}
