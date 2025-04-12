package createPvz

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/metrics"
	"avito-intern/internal/models"
	"avito-intern/internal/services"
	"encoding/json"
	"errors"
	"net/http"
)

func New(service *services.PVZService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := middleware.RequireRole(r.Context(), "moderator"); err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Access denied"})
			return
		}

		var req models.PVZ
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Invalid request"})
			return
		}

		if err := service.CreatePVZ(&req); err != nil {
			if errors.Is(err, internalErrors.ErrInvalidCity) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "City not allowed"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Internal server error"})
			}
			return
		}

		metrics.PvzCreatedCount.Inc()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(req)
	}
}
