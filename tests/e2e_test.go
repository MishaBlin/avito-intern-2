package tests

import (
	"avito-intern/internal/api"
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/api/dto/request/productDto"
	"avito-intern/internal/api/dto/request/receptionDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/models"
	"avito-intern/internal/repository"
	"avito-intern/internal/services"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDatabase(t *testing.T) (*sql.DB, func(), error) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connString, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	connString += " sslmode=disable"

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	cleanup := func() {
		db.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}

	return db, cleanup, nil
}

func applyMigrations(db *sql.DB) error {
	migrationsDir := "../internal/database/migrations"

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	for _, fileName := range migrationFiles {
		filePath := filepath.Join(migrationsDir, fileName)

		migrationSQL, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		_, err = db.Exec(string(migrationSQL))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}
	}

	return nil
}

func setupTestServer(db *sql.DB) *chi.Mux {
	userRepo := repository.NewUserRepository(db)
	pvzRepo := repository.NewPVZRepository(db)
	receptionRepo := repository.NewReceptionRepository(db)
	productRepo := repository.NewProductRepository(db)

	authService := services.NewAuthService(userRepo)
	pvzService := services.NewPVZService(pvzRepo)
	receptionService := services.NewReceptionService(receptionRepo)
	productService := services.NewProductService(productRepo, receptionRepo)

	return api.SetupRouter(
		authService,
		pvzService,
		receptionService,
		productService,
	)
}

func TestE2EWorkflow(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "true" {
		t.Skip("Skipping integration test")
	}

	dbConn, cleanup, err := setupTestDatabase(t)
	require.NoError(t, err, "Failed to setup test database")
	defer cleanup()

	router := setupTestServer(dbConn)
	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	_ = registerTestUser(t, client, server.URL, "moderator@test.com", "password", "moderator")

	_ = registerTestUser(t, client, server.URL, "employee@test.com", "password", "employee")

	moderatorToken := loginTestUser(t, client, server.URL, "moderator@test.com", "password")

	pvz := createTestPVZ(t, client, server.URL, moderatorToken)

	employeeToken := loginTestUser(t, client, server.URL, "employee@test.com", "password")

	reception := createTestReception(t, client, server.URL, employeeToken, pvz.ID)

	productTypes := []string{"электроника", "одежда", "обувь"}
	for i := 0; i < 50; i++ {
		productType := productTypes[i%len(productTypes)]
		createTestProduct(t, client, server.URL, employeeToken, pvz.ID, productType)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err = dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE receptionId = $1", reception.ID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 50, count)

	closedReception := closeTestReception(t, client, server.URL, employeeToken, pvz.ID)
	assert.Equal(t, "close", closedReception.Status)

	var status string
	err = dbConn.QueryRowContext(ctx, "SELECT status FROM receptions WHERE id = $1", reception.ID).Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, "close", status)
}

func registerTestUser(t *testing.T, client *http.Client, baseURL, email, password, role string) response.UserResponse {
	registerReq := authDto.RegisterRequest{
		Email:    email,
		Password: password,
		Role:     role,
	}

	reqBody, err := json.Marshal(registerReq)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/register", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var userResp response.UserResponse
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	require.NoError(t, err)

	return userResp
}

func loginTestUser(t *testing.T, client *http.Client, baseURL, email, password string) string {
	loginReq := authDto.LoginRequest{
		Email:    email,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/login", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp response.TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	return tokenResp.Token
}

func createTestPVZ(t *testing.T, client *http.Client, baseURL, token string) models.PVZ {
	pvz := models.PVZ{
		ID:               uuid.New().String(),
		City:             "Москва",
		RegistrationDate: time.Now(),
	}

	reqBody, err := json.Marshal(pvz)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/pvz", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPVZ models.PVZ
	err = json.NewDecoder(resp.Body).Decode(&createdPVZ)
	require.NoError(t, err)

	return createdPVZ
}

func createTestReception(t *testing.T, client *http.Client, baseURL, token, pvzID string) models.Reception {
	receptionReq := receptionDto.CreateReceptionRequest{
		PVzID: pvzID,
	}

	reqBody, err := json.Marshal(receptionReq)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/receptions", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdReception models.Reception
	err = json.NewDecoder(resp.Body).Decode(&createdReception)
	require.NoError(t, err)

	return createdReception
}

func createTestProduct(t *testing.T, client *http.Client, baseURL, token, pvzID, productType string) models.Product {
	productReq := productDto.CreateProductRequest{
		PvzID: pvzID,
		Type:  productType,
	}

	reqBody, err := json.Marshal(productReq)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/products", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdProduct models.Product
	err = json.NewDecoder(resp.Body).Decode(&createdProduct)
	require.NoError(t, err)

	return createdProduct
}

func closeTestReception(t *testing.T, client *http.Client, baseURL, token, pvzID string) models.Reception {
	req, err := http.NewRequest("POST", baseURL+"/pvz/"+pvzID+"/close_last_reception", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var closedReception models.Reception
	err = json.NewDecoder(resp.Body).Decode(&closedReception)
	require.NoError(t, err)

	return closedReception
}
