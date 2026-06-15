package middleware

import (
	"net/http"
	"strings"

	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/gin-gonic/gin"
)

type RoutePermission struct {
	Prefix  string
	Methods map[string][]string // method -> allowed roles
}

var permissions = []RoutePermission{
	{Prefix: "/api/v1/users", Methods: map[string][]string{
		http.MethodGet:    {"admin"},
		http.MethodPost:   {"admin"},
		http.MethodPut:    {"admin"},
		http.MethodDelete: {"admin"},
	}},
	{Prefix: "/api/v1/customers", Methods: map[string][]string{
		http.MethodGet:    {"admin", "director", "sales_manager", "sales_rep", "marketing", "accountant"},
		http.MethodPost:   {"admin", "sales_manager", "sales_rep"},
		http.MethodPut:    {"admin", "sales_manager", "sales_rep"},
		http.MethodDelete: {"admin", "sales_manager"},
	}},
	{Prefix: "/api/v1/opportunities", Methods: map[string][]string{
		http.MethodGet:    {"admin", "director", "sales_manager", "sales_rep"},
		http.MethodPost:   {"admin", "sales_manager", "sales_rep"},
		http.MethodPut:    {"admin", "sales_manager", "sales_rep"},
		http.MethodDelete: {"admin", "sales_manager"},
	}},
	{Prefix: "/api/v1/contracts", Methods: map[string][]string{
		http.MethodGet:  {"admin", "director", "sales_manager", "sales_rep"},
		http.MethodPost: {"admin", "sales_manager", "sales_rep"},
		http.MethodPut:  {"admin", "sales_manager", "sales_rep"},
	}},
	{Prefix: "/api/v1/quotations", Methods: map[string][]string{
		http.MethodGet:  {"admin", "director", "sales_manager", "sales_rep"},
		http.MethodPost: {"admin", "sales_manager", "sales_rep"},
		http.MethodPut:  {"admin", "sales_manager", "sales_rep"},
	}},
	{Prefix: "/api/v1/shipments", Methods: map[string][]string{
		http.MethodGet:  {"admin", "director", "sales_manager", "sales_rep", "accountant"},
		http.MethodPost: {"admin", "sales_manager"},
	}},
	{Prefix: "/api/v1/invoices", Methods: map[string][]string{
		http.MethodGet:  {"admin", "director", "accountant"},
		http.MethodPost: {"admin", "accountant"},
		http.MethodPut:  {"admin", "accountant"},
	}},
	{Prefix: "/api/v1/payments", Methods: map[string][]string{
		http.MethodGet:  {"admin", "director", "accountant"},
		http.MethodPost: {"admin", "accountant"},
	}},
	{Prefix: "/api/v1/reports", Methods: map[string][]string{
		http.MethodGet: {"admin", "director", "accountant"},
	}},
	{Prefix: "/api/v1/campaigns", Methods: map[string][]string{
		http.MethodGet:  {"admin", "director", "marketing"},
		http.MethodPost: {"admin", "marketing"},
		http.MethodPut:  {"admin", "marketing"},
	}},
	{Prefix: "/api/v1/dashboard", Methods: map[string][]string{
		http.MethodGet: {"admin", "director", "sales_manager", "sales_rep", "marketing", "accountant"},
	}},
}

func RBAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		role := GetRole(c)

		if role == "admin" {
			c.Next()
			return
		}

		for _, perm := range permissions {
			if strings.HasPrefix(path, perm.Prefix) {
				allowedRoles, ok := perm.Methods[method]
				if !ok {
					allowedRoles = perm.Methods["*"]
				}
				if allowedRoles == nil {
					c.Next()
					return
				}
				for _, r := range allowedRoles {
					if r == role {
						c.Next()
						return
					}
				}
				util.Forbidden(c, "insufficient permissions for this resource")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
