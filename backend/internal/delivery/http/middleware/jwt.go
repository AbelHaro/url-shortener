package middleware

import (
	"net/http"
	"strings"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type JWTMiddleware struct {
	authService *auth.Service
}

func NewJWTMiddleware(authService *auth.Service) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
	}
}

func (m *JWTMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]
		userID, err := m.authService.ValidateAccessToken(token)
		if err != nil {
			switch err {
			case domain.ErrTokenExpired:
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Token expired",
				})
			case domain.ErrInvalidToken:
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
				})
			default:
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Authentication failed",
				})
			}
			c.Abort()
			return
		}

		c.Set("userID", userID.String())
		c.Next()
	}
}
