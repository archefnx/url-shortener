package delete

import (
	"errors"
	"log/slog"
	"net/http"
	resp "shortener/internal/lib/api/response"
	"shortener/internal/lib/logger/sl"
	"shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	DeletedCount int64 `json:"deleted_count"`
}
type URLDeleter interface {
	DeleteURL(alias string) (int64, error)
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		resURL, err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("deleted url", slog.Int64("count", resURL))
		responseOK(w, r, resURL)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, count int64) {
	render.JSON(w, r, Response{
		Response:     resp.OK(),
		DeletedCount: count,
	})
}
