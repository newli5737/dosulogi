package util

import (
	"math"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func ParsePagination(c *gin.Context) (page, limit, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset = (page - 1) * limit
	return
}

func TotalPages(total, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}

func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}

func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, msg)
}

func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, msg)
}

func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, msg)
}

func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, msg)
}

func InternalError(c *gin.Context, msg string) {
	Error(c, http.StatusInternalServerError, msg)
}

func Paginated(c *gin.Context, items interface{}, page, limit, total int) {
	v := reflect.ValueOf(items)
	if v.Kind() == reflect.Slice && v.IsNil() {
		items = reflect.MakeSlice(v.Type(), 0, 0).Interface()
	}
	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}
