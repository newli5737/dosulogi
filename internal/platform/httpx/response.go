package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func List(c *gin.Context, data interface{}, meta Meta) {
	if data == nil {
		data = []struct{}{}
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": meta})
}

func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": ErrorBody{Code: code, Message: message}})
}

func BadRequest(c *gin.Context, code, message string) { Fail(c, http.StatusBadRequest, code, message) }
func Unauthorized(c *gin.Context, message string)      { Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", message) }
func Forbidden(c *gin.Context, message string)         { Fail(c, http.StatusForbidden, "FORBIDDEN", message) }
func NotFound(c *gin.Context, code, message string)    { Fail(c, http.StatusNotFound, code, message) }
func Conflict(c *gin.Context, code, message string)    { Fail(c, http.StatusConflict, code, message) }
func Internal(c *gin.Context, message string)          { Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", message) }

func ParsePageLimit(c *gin.Context) (page, limit, offset int) {
	page = 1
	limit = 20
	if p := c.Query("page"); p != "" {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := parseInt(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	offset = (page - 1) * limit
	return
}

func parseInt(s string) (int, error) {
	var n int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, errInvalid
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}

var errInvalid = &parseError{}

type parseError struct{}

func (e *parseError) Error() string { return "invalid" }
