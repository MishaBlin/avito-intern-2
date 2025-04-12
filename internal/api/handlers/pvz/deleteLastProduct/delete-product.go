package deleteLastProduct

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/services"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func New(service *services.ProductService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := middleware.RequireRole(r.Context(), "employee"); err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Access denied"})
			return
		}
		pvzId := chi.URLParam(r, "pvzId")
		err := service.DeleteLastProduct(pvzId)

		if err != nil {
			switch {
			case errors.Is(err, internalErrors.ErrNoActiveReception):
				{
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(response.ErrorResponse{Message: "No active reception"})
					return
				}
			case errors.Is(err, internalErrors.ErrProductNotFound):
				{
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(response.ErrorResponse{Message: "No products in reception"})
					return
				}
			default:
				{
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Internal server error"})
					return
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
