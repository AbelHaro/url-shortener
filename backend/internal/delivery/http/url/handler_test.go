package url

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	idsRangesRepository "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	idsRangesService "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	urlSvc "github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const apiRoute = "/api/v1"

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

func TestHandler_Create(t *testing.T) {
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
			router.POST(apiRoute+"/shorten", h.Create)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", apiRoute+"/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Create() status = %v, want %v", w.Code, tt.wantStatus)
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

	urlStored, err := h.Service.Store("https://google.com")
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

	urlStored, err := h.Service.Store("https://google.com")
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
			router.GET(apiRoute+"/urls/:id", h.FindByID)

			req, _ := http.NewRequest("GET", apiRoute+"/urls/"+tt.id, nil)
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

	urlStored, err := h.Service.Store("https://google.com")
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
			router.DELETE(apiRoute+"/urls/:id", h.DeleteByID)

			req, _ := http.NewRequest("DELETE", apiRoute+"/urls/"+tt.id, nil)
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

	_, err = h.Service.Store("https://google.com")
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
			router.POST(apiRoute+"/urls/search", h.FindByOriginalURL)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", apiRoute+"/urls/search", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

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
