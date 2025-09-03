package remove

import (
	"log/slog"
	"money-tracker/internal/lib/api/response"
	"money-tracker/internal/lib/logger/sl"
	mwAuth "money-tracker/internal/http-server/middleware/authorization"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type TrRemover interface {
	RemoveTransaction(trID int64, userID int64) error
}

func Remove(log *slog.Logger, trRemover TrRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.transactions.remove"

		log := log.With(slog.String("op", op), middleware.GetReqID(r.Context()))

		trIDStr := chi.URLParam(r, "id")
		trID, err := strconv.ParseInt(trIDStr, 10, 64)
		if err != nil {
			log.Error("invalid transaction id", sl.Err(err))
			render.JSON(w, r, response.Error("invalid request id"))
			return
		} 

		userID, ok := r.Context().Value(mwAuth.UserIDKey).(int64)
        if !ok {
            log.Error("user id not found in context")
            render.JSON(w, r, response.Error("user not authenticated"))
            return
        }

		err = trRemover.RemoveTransaction(trID, userID)
		if err != nil {
			log.Error("failed to remove transaction", sl.Err(err))
			render.JSON(w, r, response.Error("failed to remove transaction"))
			return
		}

		log.Info("transaction removed", slog.Int64("transaction_id", trID))
		render.JSON(w, r, response.Ok())
	}
}