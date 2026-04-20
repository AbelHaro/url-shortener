package authcookie

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
	AuthCookiePath         = "/"
	AuthCookieSameSiteMode = http.SameSiteStrictMode
)

func SetAccessToken(c *gin.Context, token string, ttl time.Duration, secure bool) {
	setCookie(c, AccessTokenCookieName, token, ttl, secure)
}

func SetRefreshToken(c *gin.Context, token string, ttl time.Duration, secure bool) {
	setCookie(c, RefreshTokenCookieName, token, ttl, secure)
}

func ClearAuthCookies(c *gin.Context, secure bool) {
	clearCookie(c, AccessTokenCookieName, secure)
	clearCookie(c, RefreshTokenCookieName, secure)
}

func setCookie(c *gin.Context, name, value string, ttl time.Duration, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     AuthCookiePath,
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
		HttpOnly: true,
		Secure:   secure,
		SameSite: AuthCookieSameSiteMode,
	})
}

func clearCookie(c *gin.Context, name string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     AuthCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: AuthCookieSameSiteMode,
	})
}
