package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	authRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/auth"
	authSvc "github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	"github.com/AbelHaro/url-shortener/backend/internal/service/jwt"
	"github.com/gin-gonic/gin"
)

const apiRoute = "/api/v1"

func provideHandler() (*Handler, error) {
	repo := authRepo.NewMockRepository()

	jwtSvc := jwt.NewService("testsecret", 15*time.Minute, 7*24*time.Hour)

	svc := authSvc.NewService(repo, jwtSvc)
	return NewHandler(svc), nil
}

func TestHandler_Register(t *testing.T) {
	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	tests := []struct {
		name       string
		request    dtos.V1RegisterRequest
		response   any
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid registration",
			request: dtos.V1RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			response: dtos.V1UserResponse{
				Email: "test@example.com",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},

		{
			name: "invalid email",
			request: dtos.V1RegisterRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			response:   dtos.V1ErrorResponse{Error: "invalid request"},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "short password",
			request: dtos.V1RegisterRequest{
				Email:    "test@example.com",
				Password: "short",
			},
			response:   dtos.V1ErrorResponse{Error: "invalid request"},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "empty request",
			request:    dtos.V1RegisterRequest{},
			response:   dtos.V1ErrorResponse{Error: "invalid request"},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST(apiRoute+"/auth/register", h.Register)

			w := httptest.NewRecorder()
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPost, apiRoute+"/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			if !tt.wantError && (w.Body.String() == "" || !bytes.Contains([]byte(w.Body.String()), []byte("id")) || !bytes.Contains([]byte(w.Body.String()), []byte("email"))) {
				t.Errorf("expected user ID and email in response, got %s", w.Body.String())
			}

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	// First, register a user to test login
	_, err = h.service.Register("test@example.com", "password123")
	if err != nil {
		t.Fatalf("h.service.Register() error = %v", err)
	}

	tests := []struct {
		name       string
		request    dtos.V1LoginRequest
		response   any
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid login",
			request: dtos.V1LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			response:   dtos.V1TokenResponse{AccessToken: "mockAccessToken", RefreshToken: "mockRefreshToken"},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "invalid password",
			request: dtos.V1LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			response:   dtos.V1ErrorResponse{Error: "invalid credentials"},
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name: "non-existent user",
			request: dtos.V1LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			response:   dtos.V1ErrorResponse{Error: "invalid credentials"},
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST(apiRoute+"/auth/login", h.Login)

			w := httptest.NewRecorder()
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPost, apiRoute+"/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			if !tt.wantError && (w.Body.String() == "" || !bytes.Contains([]byte(w.Body.String()), []byte("access_token")) || !bytes.Contains([]byte(w.Body.String()), []byte("refresh_token"))) {
				t.Errorf("expected access and refresh tokens in response, got %s", w.Body.String())
			}
			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})

	}
}

func TestHandler_RefreshToken(t *testing.T) {
	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	// First, register and login a user to get a refresh token
	_, err = h.service.Register("test@example.com", "password123")
	if err != nil {
		t.Fatalf("h.service.Register() error = %v", err)
	}

	tokens, err := h.service.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("h.service.Login() error = %v", err)
	}

	tests := []struct {
		name       string
		request    dtos.V1RefreshTokenRequest
		response   any
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid refresh token",
			request: dtos.V1RefreshTokenRequest{
				RefreshToken: tokens.RefreshToken,
			},
			response:   dtos.V1TokenResponse{AccessToken: "mockAccessToken", RefreshToken: "mockRefreshToken"},
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name: "invalid refresh token",
			request: dtos.V1RefreshTokenRequest{
				RefreshToken: "invalidtoken",
			},
			response:   dtos.V1ErrorResponse{Error: "invalid refresh token"},
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST(apiRoute+"/auth/refresh", h.RefreshToken)

			w := httptest.NewRecorder()
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPost, apiRoute+"/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			if !tt.wantError && (w.Body.String() == "" || !bytes.Contains(w.Body.Bytes(), []byte("access_token")) || !bytes.Contains(w.Body.Bytes(), []byte("refresh_token"))) {
				t.Errorf("expected access and refresh tokens in response, got %s", w.Body.String())
			}
			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	h, err := provideHandler()
	if err != nil {
		t.Fatalf("provideHandler() error = %v", err)
	}

	user, err := h.service.Register("test@example.com", "password123")
	if err != nil {
		t.Fatalf("h.service.Register() error = %v", err)
	}

	_, err = h.service.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("h.service.Login() error = %v", err)
	}

	tests := []struct {
		name       string
		setUserID  bool
		wantStatus int
	}{
		{
			name:       "valid logout with userID",
			setUserID:  true,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing userID",
			setUserID:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST(apiRoute+"/auth/logout", func(c *gin.Context) {
				if tt.setUserID {
					c.Set("userID", user.ID.String())
				}
				h.Logout(c)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, apiRoute+"/auth/logout", nil)

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
