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
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	"github.com/AbelHaro/url-shortener/backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

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

// ============= HELPER FUNCTIONS =============

// cleanupDatabase truncates all tables to ensure clean state between tests
func cleanupDatabase(t *testing.T) {
	if err := testDB.Exec("TRUNCATE TABLE urls CASCADE").Error; err != nil {
		t.Fatalf("failed to truncate urls table: %v", err)
	}
	if err := testDB.Exec("TRUNCATE TABLE users CASCADE").Error; err != nil {
		t.Fatalf("failed to truncate users table: %v", err)
	}
	if err := testDB.Exec("TRUNCATE TABLE ids_ranges CASCADE").Error; err != nil {
		t.Fatalf("failed to truncate ids_ranges table: %v", err)
	}
}

// registerUserHelper registers a new user and returns the response
func registerUserHelper(email, password string) (*dtos.V1UserResponse, error) {
	registerBody := dtos.V1RegisterRequest{Email: email, Password: password}
	body, _ := json.Marshal(registerBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "http://localhost:5173/")

	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		var errResp dtos.V1ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &errResp)
		return nil, fmt.Errorf("registration failed with status %d: %s", w.Code, errResp.Error)
	}

	var userResp dtos.V1UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &userResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registration response: %w", err)
	}

	return &userResp, nil
}

// loginUserHelper logs in a user and returns the token response
func loginUserHelper(email, password string) (*dtos.V1TokenResponse, error) {
	loginBody := dtos.V1LoginRequest{Email: email, Password: password}
	body, _ := json.Marshal(loginBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "http://localhost:5173/")

	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		var errResp dtos.V1ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &errResp)
		return nil, fmt.Errorf("login failed with status %d: %s", w.Code, errResp.Error)
	}

	var tokenResp dtos.V1TokenResponse
	if err := json.Unmarshal(w.Body.Bytes(), &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	return &tokenResp, nil
}

// makeRequest performs a request without authentication
func makeRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "http://localhost:5173/")

	testRouter.ServeHTTP(w, req)
	return w
}

// makeAuthenticatedRequest performs a request with JWT token in Authorization header
func makeAuthenticatedRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Referer", "http://localhost:5173/")

	testRouter.ServeHTTP(w, req)
	return w
}

// createShortenURLHelper creates a shortened URL and returns the response
func createShortenURLHelper(originalURL string, token string) (*dtos.V1URLResponse, error) {
	createReq := dtos.V1CreateShortenRequest{OriginalUrl: originalURL}
	w := makeAuthenticatedRequest("POST", "/api/v1/shorten", createReq, token)

	if w.Code != http.StatusCreated {
		var errResp dtos.V1ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &errResp)
		return nil, fmt.Errorf("create shorten failed with status %d: %s", w.Code, errResp.Error)
	}

	var urlResp dtos.V1URLResponse
	if err := json.Unmarshal(w.Body.Bytes(), &urlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create shorten response: %w", err)
	}

	return &urlResp, nil
}

// ============= TEST FUNCTIONS =============

func Test_HealthEndpoint(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	okJson := `{"status":"ok"}`
	assert.Equal(t, okJson, w.Body.String())
}

func Test_RegisterEndpoint(t *testing.T) {
	cleanupDatabase(t)

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
			expectedStatus: http.StatusCreated,
			expectedResponseBody: &dtos.V1UserResponse{
				Email: "test@example.com",
			},
		},
		{
			name:           "Invalid email",
			method:         "POST",
			url:            "/api/v1/auth/register",
			requestBody:    dtos.V1RegisterRequest{Email: "invalid", Password: "password123"},
			expectedStatus: http.StatusBadRequest,
			expectedResponseBody: &dtos.V1ErrorResponse{
				Error: "invalid request body",
			},
		},
		{
			name:           "Short password",
			method:         "POST",
			url:            "/api/v1/auth/register",
			requestBody:    dtos.V1RegisterRequest{Email: "test@example.com", Password: "short"},
			expectedStatus: http.StatusBadRequest,
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

func Test_LoginEndpoint(t *testing.T) {
	cleanupDatabase(t)

	// First, register a user to ensure we have valid credentials for login
	_, err := registerUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	tests := []struct {
		name                 string
		method               string
		url                  string
		requestBody          dtos.V1LoginRequest
		expectedResponseBody any
		expectedStatus       int
	}{
		{
			name:   "Valid login",
			method: "POST",
			url:    "/api/v1/auth/login",
			requestBody: dtos.V1LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus:       http.StatusOK,
			expectedResponseBody: &dtos.V1TokenResponse{},
		},
		{
			name:   "Invalid password",
			method: "POST",
			url:    "/api/v1/auth/login",
			requestBody: dtos.V1LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponseBody: &dtos.V1ErrorResponse{
				Error: "invalid credentials",
			},
		},
		{
			name:   "Non-existent user",
			method: "POST",
			url:    "/api/v1/auth/login",
			requestBody: dtos.V1LoginRequest{
				Email:    "testinvalid@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponseBody: &dtos.V1ErrorResponse{
				Error: "invalid credentials",
			},
		},
		{
			name:   "Invalid login request body with missing password",
			method: "POST",
			url:    "/api/v1/auth/login",
			requestBody: dtos.V1LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponseBody: &dtos.V1ErrorResponse{
				Error: "invalid request body",
			},
		},
		{
			name:   "Invalid login request body with missing email",
			method: "POST",
			url:    "/api/v1/auth/login",
			requestBody: dtos.V1LoginRequest{
				Email:    "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
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

			switch expected := tt.expectedResponseBody.(type) {
			case *dtos.V1TokenResponse:
				var resp dtos.V1TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			case *dtos.V1ErrorResponse:
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, expected.Error, resp.Error)
			}
		})
	}
}

func Test_RefreshTokenEndpoint(t *testing.T) {
	cleanupDatabase(t)

	// Setup: Register and login user
	_, err := registerUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	loginResp, err := loginUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	tests := []struct {
		name                 string
		method               string
		url                  string
		requestBody          dtos.V1RefreshTokenRequest
		expectedResponseBody any
		expectedStatus       int
	}{
		{
			name:   "Valid refresh token",
			method: "POST",
			url:    "/api/v1/auth/refresh",
			requestBody: dtos.V1RefreshTokenRequest{
				RefreshToken: loginResp.RefreshToken,
			},
			expectedStatus:       http.StatusOK,
			expectedResponseBody: &dtos.V1TokenResponse{},
		},
		{
			name:   "Invalid refresh token",
			method: "POST",
			url:    "/api/v1/auth/refresh",
			requestBody: dtos.V1RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			expectedStatus:       http.StatusUnauthorized,
			expectedResponseBody: &dtos.V1ErrorResponse{},
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

			switch tt.expectedResponseBody.(type) {
			case *dtos.V1TokenResponse:
				var resp dtos.V1TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			case *dtos.V1ErrorResponse:
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "invalid token", resp.Error)
			}
		})
	}
}

func Test_LogoutEndpoint(t *testing.T) {
	cleanupDatabase(t)

	// Setup: Register and login user
	_, err := registerUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	loginResp, err := loginUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		url            string
		accessToken    string
		expectedStatus int
	}{
		{
			name:           "Valid logout",
			method:         "POST",
			url:            "/api/v1/auth/logout",
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid logout with missing token",
			method:         "POST",
			url:            "/api/v1/auth/logout",
			accessToken:    "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.url, nil)
			if tt.accessToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.accessToken)
			}
			req.Header.Set("Referer", "http://localhost:5173/")

			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusUnauthorized {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "Authorization header required", resp.Error)
			}
		})
	}
}

func Test_ShortenEndpoint(t *testing.T) {
	cleanupDatabase(t)

	// Setup: Register and login user
	userResp, err := registerUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	loginResp, err := loginUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	userID := uuid.MustParse(userResp.ID)

	tests := []struct {
		name                 string
		method               string
		url                  string
		requestBody          dtos.V1CreateShortenRequest
		accessToken          string
		expectedBodyResponse any
		expectedStatus       int
	}{
		{
			name:   "Valid shorten URL",
			method: "POST",
			url:    "/api/v1/shorten",
			requestBody: dtos.V1CreateShortenRequest{
				OriginalUrl: "https://www.example.com",
			},
			accessToken: loginResp.AccessToken,
			expectedBodyResponse: &dtos.V1URLResponse{
				OriginalURL: "https://www.example.com",
				UserID:      userID,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid shorten URL with missing original URL",
			method: "POST",
			url:    "/api/v1/shorten",
			requestBody: dtos.V1CreateShortenRequest{
				OriginalUrl: "",
			},
			accessToken: loginResp.AccessToken,
			expectedBodyResponse: &dtos.V1ErrorResponse{
				Error: "invalid request body",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Unauthorized shorten URL with fake token",
			method: "POST",
			url:    "/api/v1/shorten",
			requestBody: dtos.V1CreateShortenRequest{
				OriginalUrl: "https://www.example.com",
			},
			accessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30",
			expectedBodyResponse: &dtos.V1ErrorResponse{
				Error: "Invalid token",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.url, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.accessToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.accessToken)
			}
			req.Header.Set("Referer", "http://localhost:5173/")

			testRouter.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			switch expected := tt.expectedBodyResponse.(type) {
			case *dtos.V1URLResponse:
				var resp dtos.V1URLResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, expected.OriginalURL, resp.OriginalURL)
				assert.Equal(t, expected.UserID, resp.UserID)
				assert.NotEmpty(t, resp.ShortCode)
			case *dtos.V1ErrorResponse:
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, expected.Error, resp.Error)
			}
		})
	}
}

func Test_FindByShortCodeEndpoint(t *testing.T) {
	cleanupDatabase(t)

	// Setup: Register, login, and create a shortened URL
	_, err := registerUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	loginResp, err := loginUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	urlResp, err := createShortenURLHelper("https://www.example.com", loginResp.AccessToken)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		url            string
		shortCode      string
		accessToken    string
		expectedStatus int
		assertions     func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid short code",
			method:         "GET",
			url:            "/api/v1/urls/short/",
			shortCode:      urlResp.ShortCode,
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusOK,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp domain.URL
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "https://www.example.com", resp.OriginalURL)
				assert.Equal(t, urlResp.ShortCode, resp.ShortCode)
			},
		},
		{
			name:           "Invalid short code",
			method:         "GET",
			url:            "/api/v1/urls/short/",
			shortCode:      "invalidcode",
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusNotFound,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Error)
			},
		},
		{
			name:           "Missing authentication token",
			method:         "GET",
			url:            "/api/v1/urls/short/",
			shortCode:      urlResp.ShortCode,
			accessToken:    "",
			expectedStatus: http.StatusUnauthorized,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				// The middleware returns "Invalid token" when no auth header is provided
				assert.True(t, resp.Error == "Authorization header required" || resp.Error == "Invalid token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullURL := tt.url + tt.shortCode
			w := makeAuthenticatedRequest(tt.method, fullURL, nil, tt.accessToken)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.assertions != nil {
				tt.assertions(t, w)
			}
		})
	}
}

func Test_FindByIDEndpoint(t *testing.T) {
	cleanupDatabase(t)

	// Setup: Register, login, and create a shortened URL
	_, err := registerUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	loginResp, err := loginUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	urlResp, err := createShortenURLHelper("https://www.example.com", loginResp.AccessToken)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		url            string
		id             string
		accessToken    string
		expectedStatus int
		assertions     func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid URL ID",
			method:         "GET",
			url:            "/api/v1/urls/",
			id:             urlResp.ID.String(),
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusOK,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp domain.URL
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "https://www.example.com", resp.OriginalURL)
			},
		},
		{
			name:           "Invalid URL ID (non-existent)",
			method:         "GET",
			url:            "/api/v1/urls/",
			id:             uuid.New().String(),
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusNotFound,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Error)
			},
		},
		{
			name:           "Invalid UUID format returns 500 (backend behavior)",
			method:         "GET",
			url:            "/api/v1/urls/",
			id:             "invalid-uuid",
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusInternalServerError,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Error)
			},
		},
		{
			name:           "Missing authentication token",
			method:         "GET",
			url:            "/api/v1/urls/",
			id:             urlResp.ID.String(),
			accessToken:    "",
			expectedStatus: http.StatusUnauthorized,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				// The middleware returns "Invalid token" when no auth header is provided
				assert.True(t, resp.Error == "Authorization header required" || resp.Error == "Invalid token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullURL := tt.url + tt.id
			w := makeAuthenticatedRequest(tt.method, fullURL, nil, tt.accessToken)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.assertions != nil {
				tt.assertions(t, w)
			}
		})
	}
}

func Test_DeleteByIDEndpoint(t *testing.T) {
	cleanupDatabase(t)

	// Setup: Register, login, and create a shortened URL (User 1)
	_, err := registerUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	loginResp, err := loginUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	urlResp, err := createShortenURLHelper("https://www.example.com", loginResp.AccessToken)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		url            string
		id             string
		accessToken    string
		expectedStatus int
		assertions     func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid delete (owner)",
			method:         "DELETE",
			url:            "/api/v1/urls/",
			id:             urlResp.ID.String(),
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusNoContent,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				// After deletion, verify URL doesn't exist
				verifyW := makeAuthenticatedRequest("GET", "/api/v1/urls/"+urlResp.ID.String(), nil, loginResp.AccessToken)
				assert.Equal(t, http.StatusNotFound, verifyW.Code)
			},
		},
		{
			name:           "Delete non-existent URL",
			method:         "DELETE",
			url:            "/api/v1/urls/",
			id:             uuid.New().String(),
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusNotFound,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Error)
			},
		},
		{
			name:           "Missing authentication token",
			method:         "DELETE",
			url:            "/api/v1/urls/",
			id:             urlResp.ID.String(),
			accessToken:    "",
			expectedStatus: http.StatusUnauthorized,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				// The middleware returns "Invalid token" when no auth header is provided
				assert.True(t, resp.Error == "Authorization header required" || resp.Error == "Invalid token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullURL := tt.url + tt.id
			w := makeAuthenticatedRequest(tt.method, fullURL, nil, tt.accessToken)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.assertions != nil {
				tt.assertions(t, w)
			}
		})
	}
}

func Test_SearchByOriginalURLEndpoint(t *testing.T) {
	cleanupDatabase(t)

	// Setup: Register, login, and create a shortened URL
	_, err := registerUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	loginResp, err := loginUserHelper("test@example.com", "password123")
	assert.NoError(t, err)

	urlResp, err := createShortenURLHelper("https://www.example.com", loginResp.AccessToken)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		url            string
		requestBody    dtos.V1SearchByOriginalURLRequest
		accessToken    string
		expectedStatus int
		assertions     func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "Valid search by original URL",
			method: "POST",
			url:    "/api/v1/urls/search",
			requestBody: dtos.V1SearchByOriginalURLRequest{
				URL: "https://www.example.com",
			},
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusOK,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp domain.URL
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "https://www.example.com", resp.OriginalURL)
				assert.Equal(t, urlResp.ShortCode, resp.ShortCode)
			},
		},
		{
			name:   "Search for non-existent URL returns 500 (backend behavior)",
			method: "POST",
			url:    "/api/v1/urls/search",
			requestBody: dtos.V1SearchByOriginalURLRequest{
				URL: "https://nonexistent.com",
			},
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusInternalServerError,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "internal server error", resp.Error)
			},
		},
		{
			name:   "Invalid request body (missing URL)",
			method: "POST",
			url:    "/api/v1/urls/search",
			requestBody: dtos.V1SearchByOriginalURLRequest{
				URL: "",
			},
			accessToken:    loginResp.AccessToken,
			expectedStatus: http.StatusBadRequest,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "invalid request body", resp.Error)
			},
		},
		{
			name:   "Missing authentication token",
			method: "POST",
			url:    "/api/v1/urls/search",
			requestBody: dtos.V1SearchByOriginalURLRequest{
				URL: "https://www.example.com",
			},
			accessToken:    "",
			expectedStatus: http.StatusUnauthorized,
			assertions: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp dtos.V1ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				// The middleware returns "Invalid token" when no auth header is provided
				assert.True(t, resp.Error == "Authorization header required" || resp.Error == "Invalid token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := makeAuthenticatedRequest(tt.method, tt.url, tt.requestBody, tt.accessToken)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.assertions != nil {
				tt.assertions(t, w)
			}
		})
	}
}
