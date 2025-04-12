package register

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/services"
	"encoding/json"
	"errors"
	"net/http"
)

func New(service *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authDto.RegisterRequest
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

		user, err := service.RegisterUser(req)
		if err != nil {
			if errors.Is(err, internalErrors.ErrEmailExists) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Email already exists"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Error registering user"})
			}
			return
		}

		responseUser := response.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responseUser)
	}
}
