package auth

import (
	"errors"
	"net/http"

	"github.com/dosu-logi/logistics-erp/internal/middleware"
	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	resp, refresh, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCred) {
			util.Unauthorized(c, "invalid email or password")
			return
		}
		if errors.Is(err, ErrInactive) {
			util.Unauthorized(c, "account is inactive")
			return
		}
		util.InternalError(c, err.Error())
		return
	}

	secure := c.Request.TLS != nil
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", refresh, 7*24*3600, "/", "", secure, true)
	util.JSON(c, http.StatusOK, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	refresh, err := c.Cookie("refresh_token")
	if err != nil || refresh == "" {
		util.Unauthorized(c, "missing refresh token")
		return
	}
	resp, newRefresh, err := h.svc.Refresh(c.Request.Context(), refresh)
	if err != nil {
		util.Unauthorized(c, "invalid refresh token")
		return
	}
	secure := c.Request.TLS != nil
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", newRefresh, 7*24*3600, "/", "", secure, true)
	util.JSON(c, http.StatusOK, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	refresh, _ := c.Cookie("refresh_token")
	_ = h.svc.Logout(c.Request.Context(), refresh)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	util.JSON(c, http.StatusOK, gin.H{"message": "logged out"})
}

func (h *Handler) Me(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		util.Unauthorized(c, "invalid user")
		return
	}
	u, err := h.svc.Me(c.Request.Context(), userID)
	if err != nil {
		util.NotFound(c, "user not found")
		return
	}
	u.Password = ""
	util.JSON(c, http.StatusOK, u)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		util.Unauthorized(c, "invalid user")
		return
	}
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), userID, req); err != nil {
		if errors.Is(err, ErrInvalidCred) {
			util.BadRequest(c, "incorrect old password")
			return
		}
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"message": "password updated"})
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, limit, offset := util.ParsePagination(c)
	role := c.Query("role")
	users, total, err := h.svc.ListUsers(c.Request.Context(), role, limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	for i := range users {
		users[i].Password = ""
	}
	util.Paginated(c, users, page, limit, total)
}

func (h *Handler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	u, err := h.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "user not found")
		return
	}
	u.Password = ""
	util.JSON(c, http.StatusOK, u)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	u, err := h.svc.CreateUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			util.BadRequest(c, "email already exists")
			return
		}
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, u)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	u, err := h.svc.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		util.NotFound(c, "user not found")
		return
	}
	util.JSON(c, http.StatusOK, u)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.DeleteUser(c.Request.Context(), id); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
