package createReception

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/receptionDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/metrics"
	"avito-intern/internal/models"
	"avito-intern/internal/services"
	"encoding/json"
	"errors"
	"net/http"
)

func New(service *services.ReceptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := middleware.RequireRole(r.Context(), "employee"); err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Access denied"})
			return
		}

		var req receptionDto.CreateReceptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PVzID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Invalid request"})
			return
		}

		reception := models.Reception{
			PvzID:  req.PVzID,
			Status: "in_progress",
		}

		err := service.CreateReception(&reception)
		if err != nil {
			if errors.Is(err, internalErrors.ErrActiveReceptionExists) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Active reception exists"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Internal server error"})
			}
			return
		}
		metrics.ReceptionCreatedCount.Inc()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(reception)
	}
}
