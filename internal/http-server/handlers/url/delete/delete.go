package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/xslog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5"
)

type URLDeleter interface {
	DeleteURLByAlias(ctx context.Context, alias string) (int, error)
}

type Response struct {
	response.Response
	Alias     string
	DeletedId int
}

const (
	ErrNothingToDelete = "nothing to delete"
)

func New(ctx context.Context, log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operationPlace = "handlers.url.delete.New"
		log = log.With(
			slog.String("op", operationPlace),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("empty alias"))
			return
		}

		deletedId, err := urlDeleter.DeleteURLByAlias(ctx, alias)

		if errors.Is(err, pgx.ErrNoRows) {
			log.Error("no url on", "alias", alias)
			render.JSON(w, r, response.Error("nothing to delete"))
			return
		}

		if err != nil {
			log.Error("failed to delete row", xslog.Err(err))
			render.JSON(w, r, response.Error("failed to delete row"))
			return
		}

		log.Info("success delete row by alias", "alias", alias, "deleted_id", deletedId)
		render.JSON(w, r, Response{
			Response:  response.OK(),
			Alias:     alias,
			DeletedId: deletedId,
		})
	}
}
