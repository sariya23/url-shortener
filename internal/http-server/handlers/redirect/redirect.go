package redirect

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/xslog"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURLByAlias(ctx context.Context, alias string) (string, error)
}

const (
	ErrMsgGetURL          = "failed to get URL"
	ErrMsgRedirectNoAlias = "no url on this alias"
)

func New(ctx context.Context, log *slog.Logger, getURL URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operationPlace = "handlers.redirect.New"
		log = log.With(
			slog.String("op", operationPlace),
			slog.String("requies_id", middleware.GetReqID(r.Context())),
		)
		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("empty alias"))
			return
		}

		url, err := getURL.GetURLByAlias(ctx, alias)

		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("no url on this alias", "alias", alias)
			render.JSON(w, r, response.Error(ErrMsgRedirectNoAlias))
			return
		}

		if err != nil {
			log.Error(ErrMsgGetURL, xslog.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("find url by alias", "alias", alias)
		http.Redirect(w, r, url, http.StatusFound)

	}
}
