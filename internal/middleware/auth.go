package middleware

import (
	"net/http"
	"strings"

	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "user_id"
	ContextRole   = "role"
)

func Auth(jwtMgr *util.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			util.Unauthorized(c, "missing or invalid authorization header")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtMgr.ParseAccess(token)
		if err != nil {
			util.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	v, _ := c.Get(ContextUserID)
	s, _ := v.(string)
	return s
}

func GetRole(c *gin.Context) string {
	v, _ := c.Get(ContextRole)
	s, _ := v.(string)
	return s
}

func OptionalAuth(jwtMgr *util.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			token := strings.TrimPrefix(header, "Bearer ")
			if claims, err := jwtMgr.ParseAccess(token); err == nil {
				c.Set(ContextUserID, claims.UserID)
				c.Set(ContextRole, claims.Role)
			}
		}
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role := GetRole(c)
		if !allowed[role] {
			util.Forbidden(c, "insufficient permissions")
			c.Abort()
			return
		}
		c.Next()
	}
}

func DirectorReadOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetRole(c)
		if role == "director" && c.Request.Method != http.MethodGet {
			util.Forbidden(c, "director has read-only access")
			c.Abort()
			return
		}
		c.Next()
	}
}
