package auth

import (
	"log/slog"
	"net/http"

	"money-tracker/internal/lib/api/response"
	"money-tracker/internal/lib/logger/sl"
	"money-tracker/internal/model"
	
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type UserSaver interface {
	SaveUser(email string, password string) (int64, error)
	GetUserByEmail(email string) (*model.User, error)
}

type RegisterResponse struct {
	response.Response
	UserID int64 `json:"user_id"`
}

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(log *slog.Logger, userSaver UserSaver, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.Register"

		log := slog.With(slog.String("op", op), middleware.GetReqID(r.Context()))

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode req body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode req"))
			return
		}

		if req.Email == "" || req.Password == "" {
			log.Error("empty email or password")
			render.JSON(w, r, response.Error("email and password are required"))
			return
		}

		if len(req.Password) < 8 {
			log.Error("password too short")
			render.JSON(w, r, response.Error("password must contain at least 8 characters"))
			return
		}

		userID, err := userSaver.SaveUser(req.Email, req.Password)
		if err != nil {
			log.Error("failed to save a new user", sl.Err(err))
			render.JSON(w, r, response.Error("failed to create a user"))
			return
		}

		log.Info("user registered", slog.Int64("user_id", userID))

		render.JSON(w, r, RegisterResponse{
			Response: response.Ok(),
			UserID:   userID,
		})
	}
}
