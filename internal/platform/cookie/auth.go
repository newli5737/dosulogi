package cookie

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessToken  = "access_token"
	RefreshToken = "refresh_token"
)

type AuthConfig struct {
	Domain string
}

func isSecure(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	return c.GetHeader("X-Forwarded-Proto") == "https"
}

func SetAuthCookies(c *gin.Context, cfg AuthConfig, access string, accessTTL time.Duration, refresh string, refreshTTL time.Duration) {
	secure := isSecure(c)
	domain := cfg.Domain
	maxAccess := int(accessTTL.Seconds())
	maxRefresh := int(refreshTTL.Seconds())
	if domain != "" || secure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie(AccessToken, access, maxAccess, "/", domain, secure, true)
	c.SetCookie(RefreshToken, refresh, maxRefresh, "/", domain, secure, true)
}

func ClearAuthCookies(c *gin.Context, cfg AuthConfig) {
	secure := isSecure(c)
	if cfg.Domain != "" || secure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie(AccessToken, "", -1, "/", cfg.Domain, secure, true)
	c.SetCookie(RefreshToken, "", -1, "/", cfg.Domain, secure, true)
}

func AccessFromRequest(c *gin.Context) string {
	v, err := c.Cookie(AccessToken)
	if err != nil {
		return ""
	}
	return v
}
