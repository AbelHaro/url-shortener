package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	counterRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/counter"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	urlSvc "github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ApiRoute = "/api/v1"

func provideHandler() (*URLHandler, error) {
	repo := url.NewMockRepository()
	counterRepoInstance := counterRepo.NewMockRepository()
	counterSvcInstance, err := counterSvc.NewService(counterRepoInstance)
	if err != nil {
		return nil, err
	}
	svc := urlSvc.NewService(repo, counterSvcInstance)
	return NewURLHandler(svc), nil
}

func TestURLHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "valid url",
			body:       map[string]string{"original_url": "https://google.com"},
			wantStatus: http.StatusCreated,
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
			router.POST(ApiRoute+"/shorten", h.Create)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", ApiRoute+"/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handler.Create() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestURLHandler_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	urlStored, err := h.service.Store("https://google.com")
	if err != nil {
		t.Fatalf("service.Store() error = %v", err)
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
				t.Errorf("handler.Redirect() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestURLHandler_FindByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	urlStored, err := h.service.Store("https://google.com")
	if err != nil {
		t.Fatalf("service.Store() error = %v", err)
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
			router.GET(ApiRoute+"/urls/:id", h.FindByID)

			req, _ := http.NewRequest("GET", ApiRoute+"/urls/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handler.FindByID() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestURLHandler_DeleteByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	urlStored, err := h.service.Store("https://google.com")
	if err != nil {
		t.Fatalf("service.Store() error = %v", err)
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
			router.DELETE(ApiRoute+"/urls/:id", h.DeleteByID)

			req, _ := http.NewRequest("DELETE", ApiRoute+"/urls/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handler.DeleteByID() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestURLHandler_FindByOriginalURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	_, err = h.service.Store("https://google.com")
	if err != nil {
		t.Fatalf("service.Store() error = %v", err)
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
			router.POST(ApiRoute+"/urls/search", h.FindByOriginalURL)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", ApiRoute+"/urls/search", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handler.FindByOriginalURL() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestURLHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	router := gin.New()
	router.GET(ApiRoute+"/health", h.Health)

	req, _ := http.NewRequest("GET", ApiRoute+"/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handler.Health() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response HealthResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		return
	}
	if response.Status != "ok" {
		t.Errorf("handler.Health() response = %v, want %v", response.Status, "ok")
	}
}

func TestURLHandler_HandleError(t *testing.T) {
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
