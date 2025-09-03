package auth

import (
	"log/slog"
	"net/http"
	"time"

	"money-tracker/internal/lib/api/response"
	"money-tracker/internal/lib/jwt"
	"money-tracker/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	response.Response
	Token  string `json:"token"`
	UserID int64  `json:"user_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(log *slog.Logger, userSaver UserSaver, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.Login"

		log := slog.With(slog.String("op", op), middleware.GetReqID(r.Context()))

		var req LoginRequest

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode req body")
			render.JSON(w, r, response.Error("failed to decode req"))
			return
		}

		user, err := userSaver.GetUserByEmail(req.Email)
		if err != nil {
			log.Error("user not found")
			render.JSON(w, r, response.Error("invalid credentials"))
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			log.Error("invalid password", sl.Err(err))
			render.JSON(w, r, response.Error("invalid credentials"))
			return
		}

		token, err := jwt.GenerateToken(user.ID, jwtSecret, 24*time.Hour)
		if err != nil {
			log.Error("failed to generate token", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		log.Info("user logged in", slog.Int64("user_id", user.ID))

		render.JSON(w, r, LoginResponse{
			Response: response.Ok(),
			Token:    token,
			UserID:   user.ID,
		})
	}
}
