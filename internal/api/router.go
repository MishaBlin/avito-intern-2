package api

import (
	"avito-intern/internal/api/handlers/auth/dummyLogin"
	"avito-intern/internal/api/handlers/auth/login"
	"avito-intern/internal/api/handlers/auth/register"
	"avito-intern/internal/api/handlers/product/createProduct"
	"avito-intern/internal/api/handlers/pvz/closeReception"
	"avito-intern/internal/api/handlers/pvz/createPvz"
	"avito-intern/internal/api/handlers/pvz/deleteLastProduct"
	"avito-intern/internal/api/handlers/pvz/listPvz"
	"avito-intern/internal/api/handlers/reception/createReception"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/services"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(
	authService *services.AuthService,
	pvzService *services.PVZService,
	receptionService *services.ReceptionService,
	productService *services.ProductService,
) *chi.Mux {
	router := chi.NewRouter()
	router.Use(chimw.Logger)
	router.Use(chimw.Recoverer)
	router.Use(chimw.URLFormat)
	router.Use(middleware.MetricsMiddleware)

	router.Handle("/metrics", promhttp.Handler())

	router.Post("/register", register.New(authService))
	router.Post("/login", login.New(authService))
	router.Post("/dummyLogin", dummyLogin.New())

	router.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Post("/pvz", createPvz.New(pvzService))
		r.Get("/pvz", listPvz.New(pvzService))
		r.Post("/receptions", createReception.New(receptionService))
		r.Post("/pvz/{pvzId}/close_last_reception", closeReception.New(receptionService))
		r.Post("/pvz/{pvzId}/delete_last_product", deleteLastProduct.New(productService))
		r.Post("/products", createProduct.New(productService))
	})

	return router
}
