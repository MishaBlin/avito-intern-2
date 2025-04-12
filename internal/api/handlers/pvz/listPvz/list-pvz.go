package listPvz

import (
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/services"
	"encoding/json"
	"log"
	"net/http"
)

func New(service *services.PVZService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, err := middleware.GetUserFromContext(r.Context())
		if err != nil || (user.Role != "employee" && user.Role != "moderator") {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Access denied"})
			return
		}

		startDateStr := r.URL.Query().Get("startDate")
		endDateStr := r.URL.Query().Get("endDate")
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")

		pvzs, err := service.ListPVZ(limitStr, pageStr, startDateStr, endDateStr)
		log.Println(err)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response.ErrorResponse{Message: "Internal server error"})
			return
		}
		json.NewEncoder(w).Encode(pvzs)
	}
}
