package save

import (
	"log/slog"
	"net/http"
	"time"

	mwAuth "money-tracker/internal/http-server/middleware/authorization"
	"money-tracker/internal/lib/api/response"
	"money-tracker/internal/lib/logger/sl"
	"money-tracker/internal/model"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	Tr model.Trasaction `json:"transaction"`
}

type TrSaver interface {
	SaveTransaction(tr model.Trasaction) (int64, error)
}

func New(log *slog.Logger, trSaver TrSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.transactions.New"

		log := log.With(slog.String("op", op), middleware.GetReqID(r.Context()))

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode req body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode req"))
			return
		}

		log.Info("req body decoded", slog.Any("requset", req))

		userID, ok := r.Context().Value(mwAuth.UserIDKey).(int64)
		if !ok {
			log.Error("user id not found in context")
			render.JSON(w, r, response.Error("user not authenticated"))
			return
		}

		tr := model.Trasaction{
			UserID:      userID,
			Amount:      req.Tr.Amount,
			Category:    req.Tr.Category,
			Description: req.Tr.Description,
			Date:        time.Now(),
		}

		id, err := trSaver.SaveTransaction(tr)
		if err != nil {
			log.Error("failed to save transaction", sl.Err(err))
			render.JSON(w, r, "failed to save transaction")
			return
		}

		log.Info("transaction saved", slog.Int64("id", id))

		render.JSON(w, r, response.Ok())
	}
}
