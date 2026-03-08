package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler()

	router := gin.New()
	router.GET("/health", h.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health() status = %v, want %v", w.Code, http.StatusOK)
	}

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Health() response.Status = %v, want %v", response.Status, "ok")
	}
}

func TestHandler_Health_ContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler()

	router := gin.New()
	router.GET("/health", h.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Health() Content-Type = %v, want %v", contentType, "application/json; charset=utf-8")
	}
}
