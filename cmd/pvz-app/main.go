package main

import (
	"avito-intern/internal/api"
	"avito-intern/internal/database"
	"avito-intern/internal/repository"
	"avito-intern/internal/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "admin")
	dbName := getEnv("DB_NAME", "pvz-db")

	dbAddress := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	dbConn, err := database.NewPostgres(dbAddress)

	if err != nil {
		log.Fatal("Could not connect to database: ", err)
	}

	userRepo := repository.NewUserRepository(dbConn)
	pvzRepo := repository.NewPVZRepository(dbConn)
	receptionRepo := repository.NewReceptionRepository(dbConn)
	productRepo := repository.NewProductRepository(dbConn)

	authService := services.NewAuthService(userRepo)
	pvzService := services.NewPVZService(pvzRepo)
	receptionService := services.NewReceptionService(receptionRepo)
	productService := services.NewProductService(productRepo, receptionRepo)

	router := api.SetupRouter(
		authService,
		pvzService,
		receptionService,
		productService,
	)

	go func() {
		metricsRouter := http.NewServeMux()
		metricsRouter.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics server is running on :9000")
		if err := http.ListenAndServe(":9000", metricsRouter); err != nil {
			log.Printf("Failed to start metrics server: %v", err)
		}
	}()

	port := getEnv("APP_PORT", "8080")
	log.Printf("Server is running on :%s", port)
	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
