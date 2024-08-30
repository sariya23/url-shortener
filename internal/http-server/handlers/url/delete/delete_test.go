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
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteQuerySuccess(t *testing.T) {
	cases := []struct {
		caseName   string
		alias      string
		deletedId  int
		respStatus string
		respError  string
		mockError  error
	}{
		{
			caseName:   "Success delete row",
			alias:      "qwe",
			deletedId:  1,
			respStatus: response.StatusOK,
		},
		{
			caseName:   "No row with this alias",
			alias:      "qwe",
			deletedId:  0,
			respStatus: response.StatusError,
			respError:  "nothing to delete",
			mockError:  pgx.ErrNoRows,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			urlDeleterMock := mocks.NewURLDeleter(t)
			urlDeleterMock.On("DeleteURLByAlias", ctx, tc.alias).Return(tc.deletedId, tc.mockError)
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

			assert.Equal(t, tc.respStatus, respBody.Status)
			assert.Equal(t, tc.deletedId, respBody.DeletedId)
			assert.Equal(t, tc.respError, respBody.Error)
		})
	}
}
