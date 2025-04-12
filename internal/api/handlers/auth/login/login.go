package login

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/models"
	"encoding/json"
	"errors"
	"net/http"
)

type AuthService interface {
	RegisterUser(req authDto.RegisterRequest) (*models.User, error)
	AuthenticateUser(req authDto.LoginRequest) (string, error)
}

func New(service AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authDto.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		token, err := service.AuthenticateUser(req)
		if err != nil {
			if errors.Is(err, internalErrors.ErrInvalidCredentials) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Invalid credentials"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Could not generate token"})
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response.TokenResponse{Token: token})
	}
}
