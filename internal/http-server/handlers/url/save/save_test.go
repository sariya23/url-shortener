package save_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/save/mocks"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogpretty/slogdiscard"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		caseName    string
		urlToSave   string
		aliasForURL string
		responseErr string
		mockErr     error
	}{
		{
			caseName:    "Success save",
			urlToSave:   "http://test.ru",
			aliasForURL: "suc",
		},
		{
			caseName:    "Long url",
			urlToSave:   "http://chart.apis.google.com/chart?chs=500x500&chma=0,0,100,100&cht=p&chco=FF0000%2CFFFF00%7CFF8000%2C00FF00%7C00FF00%2C0000FF&chd=t%3A122%2C42%2C17%2C10%2C8%2C7%2C7%2C7%2C7%2C6%2C6%2C6%2C6%2C5%2C5&chl=122%7C42%7C17%7C10%7C8%7C7%7C7%7C7%7C7%7C6%7C6%7C6%7C6%7C5%7C5&chdl=android%7Cjava%7Cstack-trace%7Cbroadcastreceiver%7Candroid-ndk%7Cuser-agent%7Candroid-webview%7Cwebview%7Cbackground%7Cmultithreading%7Candroid-source%7Csms%7Cadb%7Csollections%7Cactivity",
			aliasForURL: "sucv2",
		},
		{
			caseName:  "No allias",
			urlToSave: "http://test.ru",
		},
		{
			caseName:    "empty url",
			urlToSave:   "",
			aliasForURL: "empty url",
			responseErr: fmt.Sprintf("%s %s", response.ErrMSgMissingRequiredField, "URL"),
		},
		{
			caseName:    "SaveURL error",
			urlToSave:   "http://qwe.ru",
			responseErr: save.ErrMsgFailedAddUrl,
			mockErr:     errors.New("unexpected error"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.caseName, func(t *testing.T) {
			ctx := context.Background()
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			if testCase.responseErr == "" || testCase.mockErr != nil {
				urlSaverMock.On("SaveURL", ctx, testCase.urlToSave, mock.AnythingOfType("string")).
					Return(1, testCase.mockErr).
					Once()
			}

			handler := save.New(ctx, slogdiscard.NewDiscardLogger(), urlSaverMock)
			dataToRequest := fmt.Sprintf(`{"url":"%s", "alias":"%s"}`, testCase.urlToSave, testCase.aliasForURL)

			request, err := http.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(dataToRequest)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, request)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			var response save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &response))
			require.Equal(t, testCase.responseErr, response.Error)
		})
	}
}
