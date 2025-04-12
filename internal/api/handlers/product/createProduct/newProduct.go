package createProduct

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/productDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/metrics"
	"avito-intern/internal/services"
	"encoding/json"
	"errors"
	"net/http"
)

func New(productService *services.ProductService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := middleware.RequireRole(r.Context(), "employee"); err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Access denied"})
			return
		}

		var req productDto.CreateProductRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PvzID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Invalid request"})
			return
		}

		product, err := productService.AddProduct(&req)
		if err != nil {
			switch {
			case errors.Is(err, internalErrors.ErrNoActiveReception):
				{
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(response.ErrorResponse{Message: "No active reception"})
					return
				}
			case errors.Is(err, internalErrors.ErrInvalidProductType):
				{
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Invalid product type"})
					return
				}
			default:
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Invalid product type"})
				return
			}
		}

		metrics.ProductAddedCount.Inc()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(product)
	}
}
