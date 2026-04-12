package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/config"
	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	"github.com/AbelHaro/url-shortener/backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

const dummyPostgresPassword = "password"
const dummyPostgresDB = "url_shortener_test"

var testDB *gorm.DB
var testRouter *gin.Engine
var appConfig *config.AppConfig

func TestMain(m *testing.M) {
	ctx := context.Background()

	appConfig = &config.AppConfig{
		DBConfig: config.DBConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "user",
			Password: "password",
			DBName:   "url_shortener_test",
		},
		RangeConfig: config.RangeConfig{
			RangeSize:   1000,
			RangeOffset: 100,
		},
		Host:       "localhost",
		Port:       "8080",
		JWTSecret:  "test-secret-key",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 168 * time.Hour,
		Production: false,
	}

	// Start PostgreSQL container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:18",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": appConfig.DBConfig.Password,
				"POSTGRES_DB":       appConfig.DBConfig.DBName,
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})

	if err != nil {
		panic(fmt.Sprintf("Could not start container: %s", err))
	}

	// Get the database connection string from the container
	host, err := container.Host(ctx)
	if err != nil {
		panic(fmt.Sprintf("Could not get container host: %s", err))
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(fmt.Sprintf("Could not get container port: %s", err))
	}

	// Set database environment variables for the database connection
	dsn := fmt.Sprintf("postgres://postgres:%s@%s:%s/%s?sslmode=disable",
		appConfig.DBConfig.Password, host, port.Port(), appConfig.DBConfig.DBName)

	os.Setenv("DATABASE_URL", dsn)
	os.Setenv("JWT_SECRET", appConfig.JWTSecret)

	// Initialize database connection
	var dbErr error
	testDB, dbErr = database.NewDBFromDSN(dsn)
	if dbErr != nil {
		panic(fmt.Sprintf("Could not connect to database: %s", dbErr))
	}

	// Initialize router with all configured dependencies
	var routerErr error
	testRouter, routerErr = NewConfiguredRouter(testDB, appConfig)
	if routerErr != nil {
		panic(fmt.Sprintf("Could not configure router: %s", routerErr))
	}

	code := m.Run()

	container.Terminate(ctx)
	os.Exit(code)
}

func Test_HealthEndpoint(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	okJson := `{"status":"ok"}`
	assert.Equal(t, okJson, w.Body.String())
}

func Test_RegisterEndpoint(t *testing.T) {

	tests := []struct {
		name                 string
		method               string
		url                  string
		requestBody          dtos.V1RegisterRequest
		expectedResponseBody any
		expectedStatus       int
	}{
		{
			name:           "Valid registration",
			method:         "POST",
			url:            "/api/v1/auth/register",
			requestBody:    dtos.V1RegisterRequest{Email: "test@example.com", Password: "password123"},
			expectedStatus: 201,
			expectedResponseBody: &dtos.V1UserResponse{
				Email: "test@example.com",
			},
		},
		{
			name:           "Invalid email",
			method:         "POST",
			url:            "/api/v1/auth/register",
			requestBody:    dtos.V1RegisterRequest{Email: "invalid", Password: "password123"},
			expectedStatus: 400,
			expectedResponseBody: &dtos.V1ErrorResponse{
				Error: "invalid request body",
			},
		},
		{
			name:           "Short password",
			method:         "POST",
			url:            "/api/v1/auth/register",
			requestBody:    dtos.V1RegisterRequest{Email: "test@example.com", Password: "short"},
			expectedStatus: 400,
			expectedResponseBody: &dtos.V1ErrorResponse{
				Error: "invalid request body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.url, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Referer", "http://localhost:5173/")

			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Unmarshal response based on expected type
			switch expected := tt.expectedResponseBody.(type) {
			case *dtos.V1UserResponse:
				var resp dtos.V1UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				if expected.Email != "" {
					assert.Equal(t, expected.Email, resp.Email)
				}
			case *dtos.V1ErrorResponse:
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, expected.Error, resp.Error)
			}
		})
	}
}
