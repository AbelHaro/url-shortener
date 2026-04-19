package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	authcookie "github.com/AbelHaro/url-shortener/backend/internal/delivery/http/authcookie"
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type JWTMiddleware struct {
	authService  *auth.Service
	secureCookie bool
}

func NewJWTMiddleware(authService *auth.Service, secureCookie bool) *JWTMiddleware {
	return &JWTMiddleware{authService: authService, secureCookie: secureCookie}
}

func (m *JWTMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		fmt.Printf("JWT Middleware: Authenticating request...\n")
		fmt.Printf("Request URL: %s, Method: %s, Cookies: %v\n", c.Request.URL.Path, c.Request.Method, c.Request.Cookies())

		token := m.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			c.Abort()
			return
		}

		claims, err := m.authService.ValidateAccessTokenClaims(token)
		if err != nil {
			switch err {
			case domain.ErrTokenExpired:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			case domain.ErrInvalidToken:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			}
			c.Abort()
			return
		}

		userID, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if m.shouldRotate(claims) {
			if user, err := m.authService.Session(userID.String()); err == nil {
				if newToken, err := m.authService.IssueAccessToken(user.ID, user.Email); err == nil {
					authcookie.SetAccessToken(c, newToken, m.authService.AccessTTL(), m.secureCookie)
				}
			}
		}

		c.Set("userID", userID.String())
		c.Next()
	}
}

func (m *JWTMiddleware) extractToken(c *gin.Context) string {
	if cookie, err := c.Cookie(authcookie.AccessTokenCookieName); err == nil && cookie != "" {
		return cookie
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

func (m *JWTMiddleware) shouldRotate(claims map[string]any) bool {
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false
	}

	remaining := time.Until(time.Unix(int64(exp), 0))
	return remaining > 0 && remaining <= 10*time.Minute
}
