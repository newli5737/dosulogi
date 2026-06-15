package util

import "github.com/gin-gonic/gin"

// JSONList ensures empty Go slices serialize as [] not null in JSON.
func JSONList[T any](c *gin.Context, code int, items []T) {
	if items == nil {
		items = []T{}
	}
	c.JSON(code, items)
}
