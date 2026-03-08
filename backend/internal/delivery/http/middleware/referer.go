package middleware

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

var allowedOrigins = []string{
	"http://localhost:5173/",
	"https://url-shortener.abelharo.me/",
}

type RefererMiddleware struct{}

func NewRefererMiddleware() *RefererMiddleware {
	return &RefererMiddleware{}
}

func (m *RefererMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		referer := c.GetHeader("Referer")
		if referer == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
			return
		}

		valid := false
		refererURL, err := url.Parse(referer)
		if err == nil {
			refererHost := refererURL.Host
			for _, origin := range allowedOrigins {
				originURL, parseErr := url.Parse(origin)
				if parseErr == nil && refererHost == originURL.Host {
					valid = true
					break
				}
			}
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
			return
		}

		c.Next()
	}
}
