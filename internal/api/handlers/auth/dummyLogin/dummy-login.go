package dummyLogin

import (
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/models"
	"avito-intern/internal/utils"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authDto.DummyLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Invalid request"})
			return
		}
		if req.Role != "employee" && req.Role != "moderator" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Invalid role"})
			return
		}

		user := models.User{
			ID:    uuid.New().String(),
			Role:  req.Role,
			Email: "dummy@example.com",
		}
		token, err := utils.GenerateJWT(user.ID, user.Role)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Could not generate token"})
			return
		}
		json.NewEncoder(w).Encode(response.TokenResponse{Token: token})
	}
}
