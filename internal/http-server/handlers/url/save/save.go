package save

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/xslog"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int, error)
}

func New(ctx context.Context, log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operationPlace = "handlers.url.save.New"
		log = log.With(
			slog.String("op", operationPlace),
			slog.String("requies_id", middleware.GetReqID(r.Context())),
		)

		var request Request

		err := render.DecodeJSON(r.Body, &request)
		if err != nil {
			log.Error("failed to decode request body", xslog.Err(err))
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", request))

		err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

		if err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("invalid request data", xslog.Err(err))
			render.JSON(w, r, response.ValidationError(validationErrors))
			return
		}

		alias := request.Alias
		if alias == "" {
			alias = random.NewRandomString(random.DefaultStringLen)
		}

		id, err := urlSaver.SaveURL(ctx, request.URL, alias)

		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", "url", request.URL)
			render.JSON(w, r, response.Error("url already exists"))
			return
		}

		if err != nil {
			log.Error("failed to add url", xslog.Err(err))
			render.JSON(w, r, response.Error("failed to add url"))
			return
		}

		log.Info("url added", slog.Int("id", id))
		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})
	}
}
