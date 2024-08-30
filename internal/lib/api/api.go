package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"url-shortener/internal/lib/api/response"
)

var (
	ErrInvalidStatusCode = errors.New("invalid status code")
)

// GetRedirect returns the final URL after redirection.
func GetRedirect(url string) (string, error) {
	const op = "api.GetRedirect"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // stop after 1st redirect
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("%s: %w: %d", op, ErrInvalidStatusCode, resp.StatusCode)
	}

	return resp.Header.Get("Location"), nil
}

func SendGet(url string) (*response.Response, error) {
	client := http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	data, _ := io.ReadAll(resp.Body)
	var res response.Response
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
