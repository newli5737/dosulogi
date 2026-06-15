package middleware

import (
	"net/http"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/platform/cache"
	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/gin-gonic/gin"
)

func RateLimit(store *cache.Store, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil || !store.Available() {
			c.Next()
			return
		}
		userID := GetUserID(c)
		if userID == "" {
			userID = c.ClientIP()
		}
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		ok, err := store.AllowRate(c.Request.Context(), userID, route, limit, time.Minute)
		if err != nil {
			c.Next()
			return
		}
		if !ok {
			util.JSON(c, http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}
