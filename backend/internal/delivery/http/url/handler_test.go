package url

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/middleware"
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	authRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/auth"
	idsRangesRepository "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	authSvc "github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	idsRangesService "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	jwtSvc "github.com/AbelHaro/url-shortener/backend/internal/service/jwt"
	urlSvc "github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const apiRoute = "/api/v1"
const testJWTSecret = "test-secret-key-for-jwt-testing"

func generateTestToken(userID uuid.UUID) string {
	jwtService := jwtSvc.NewService(testJWTSecret, 15*time.Minute, 24*time.Hour)
	token, err := jwtService.GenerateAccessToken(userID, "test@example.com")
	if err != nil {
		panic(err)
	}
	return token
}

func provideHandler() (*Handler, error) {
	idsRangesRepository := idsRangesRepository.NewMockRepository()
	idsRangesService := idsRangesService.NewService(idsRangesRepository)
	counterSvcInstance, err := counterSvc.NewService(idsRangesService)
	if err != nil {
		return nil, err
	}

	urlRepository := url.NewMockRepository()
	svc := urlSvc.NewService(urlRepository, counterSvcInstance)
	return NewHandler(svc), nil
}

func provideJWTMiddleware() *middleware.JWTMiddleware {
	authRepository := authRepo.NewMockRepository()
	jwtService := jwtSvc.NewService(testJWTSecret, 15*time.Minute, 24*time.Hour)
	authService := authSvc.NewService(authRepository, jwtService)
	return middleware.NewJWTMiddleware(authService)
}

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	jwtMiddleware := provideJWTMiddleware()

	userID := uuid.New()
	token := generateTestToken(userID)

	tests := []struct {
		name       string
		body       dtos.V1CreateShortenRequest
		wantStatus int
	}{
		{
			name:       "valid url",
			body:       dtos.V1CreateShortenRequest{OriginalUrl: "https://google.com"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid request",
			body:       dtos.V1CreateShortenRequest{OriginalUrl: "invalid-url"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			// Apply JWT middleware to the route
			router.POST(apiRoute+"/shorten", jwtMiddleware.Authenticate(), h.Create)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", apiRoute+"/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Sended body: %v", string(body))
				t.Errorf("Create() status = %v, want %v", w.Code, tt.wantStatus)
				t.Errorf("Message = %v", w.Body.String())
			}
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	urlStored, err := h.Service.Store("https://google.com", uuid.New())
	if err != nil {
		t.Fatalf("Service.Store() error = %v", err)
	}

	tests := []struct {
		name       string
		shortCode  string
		wantStatus int
	}{
		{
			name:       "existing short URL",
			shortCode:  urlStored.ShortCode,
			wantStatus: http.StatusMovedPermanently,
		},
		{
			name:       "not found short URL",
			shortCode:  "notfound",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/:shortURL", h.Redirect)

			req, _ := http.NewRequest("GET", "/"+tt.shortCode, nil)
			req.Header.Set("Referer", "http://localhost:5173/")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Redirect() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandler_FindByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	jwtMiddleware := provideJWTMiddleware()

	userID := uuid.New()
	token := generateTestToken(userID)

	urlStored, err := h.Service.Store("https://google.com", userID)
	if err != nil {
		t.Fatalf("Service.Store() error = %v", err)
	}

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "existing id",
			id:         urlStored.ID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found id",
			id:         "550e8400-e29b-41d4-a716-446655440000",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id",
			id:         "invalid",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET(apiRoute+"/urls/:id", jwtMiddleware.Authenticate(), h.FindByID)

			req, _ := http.NewRequest("GET", apiRoute+"/urls/"+tt.id, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("FindByID() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandler_DeleteByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	jwtMiddleware := provideJWTMiddleware()

	userID := uuid.New()
	token := generateTestToken(userID)

	urlStored, err := h.Service.Store("https://google.com", userID)
	if err != nil {
		t.Fatalf("Service.Store() error = %v", err)
	}

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "existing id",
			id:         urlStored.ID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "not found id",
			id:         uuid.New().String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.DELETE(apiRoute+"/urls/:id", jwtMiddleware.Authenticate(), h.DeleteByID)

			req, _ := http.NewRequest("DELETE", apiRoute+"/urls/"+tt.id, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("DeleteByID() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandler_FindByOriginalURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	jwtMiddleware := provideJWTMiddleware()

	userID := uuid.New()
	token := generateTestToken(userID)

	_, err = h.Service.Store("https://google.com", userID)
	if err != nil {
		t.Fatalf("Service.Store() error = %v", err)
	}

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "existing url",
			body:       map[string]string{"url": "https://google.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found url",
			body:       map[string]string{"url": "https://notfound.com"},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid request",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST(apiRoute+"/urls/search", jwtMiddleware.Authenticate(), h.FindByOriginalURL)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", apiRoute+"/urls/search", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("FindByOriginalURL() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandler_HandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "ErrURLNotFound",
			err:        domain.ErrURLNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrInvalidURL",
			err:        domain.ErrInvalidURL,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "generic error",
			err:        domain.ErrInternal,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := provideHandler()
			if err != nil {
				t.Fatalf("provideHandler() error = %v", err)
			}

			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				h.handleError(c, tt.err)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleError() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
