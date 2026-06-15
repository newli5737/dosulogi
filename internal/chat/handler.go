package chat

import (
	"net/http"

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

func (h *Handler) ListAccounts(c *gin.Context) {
	items, err := h.svc.ListAccounts(c.Request.Context(), c.Query("platform"))
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var body struct {
		Platform    string `json:"platform" binding:"required"`
		Name        string `json:"name" binding:"required"`
		CookiesJSON string `json:"cookies_json"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	var createdBy *uuid.UUID
	if uid, ok := c.Get("user_id"); ok {
		if s, ok := uid.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				createdBy = &id
			}
		}
	}
	item, err := h.svc.CreateAccount(c.Request.Context(), body.Platform, body.Name, body.CookiesJSON, createdBy)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, item)
}

func (h *Handler) UpdateAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	var body struct {
		Name        *string `json:"name"`
		CookiesJSON *string `json:"cookies_json"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.UpdateAccount(c.Request.Context(), id, body.Name, body.CookiesJSON)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, item)
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.DeleteAccount(c.Request.Context(), id); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) ZaloQRLogin(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	out, err := h.svc.ZaloQRLogin(c.Request.Context(), id)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, out)
}

func (h *Handler) ZaloLoginStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	out, err := h.svc.ZaloLoginStatus(c.Request.Context(), id)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, out)
}

func (h *Handler) Inbox(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		util.BadRequest(c, "account_id required")
		return
	}
	id, err := uuid.Parse(accountID)
	if err != nil {
		util.BadRequest(c, "invalid account_id")
		return
	}
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")
	inbox, err := h.svc.FetchInbox(c.Request.Context(), id, strVal(role), strVal(userID))
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, inbox)
}

func (h *Handler) Thread(c *gin.Context) {
	accountID := c.Query("account_id")
	threadID := c.Param("threadId")
	if accountID == "" || threadID == "" {
		util.BadRequest(c, "account_id and threadId required")
		return
	}
	id, err := uuid.Parse(accountID)
	if err != nil {
		util.BadRequest(c, "invalid account_id")
		return
	}
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")
	thread, err := h.svc.FetchThread(c.Request.Context(), id, threadID, c.Query("cursor"), strVal(role), strVal(userID))
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, thread)
}

func (h *Handler) Send(c *gin.Context) {
	var body struct {
		AccountID string `json:"account_id" binding:"required"`
		ThreadID  string `json:"thread_id" binding:"required"`
		Text      string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	id, err := uuid.Parse(body.AccountID)
	if err != nil {
		util.BadRequest(c, "invalid account_id")
		return
	}
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")
	if err := h.svc.SendMessage(c.Request.Context(), id, body.ThreadID, body.Text, strVal(role), strVal(userID)); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"status": "sent"})
}

func (h *Handler) GetConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	conv, err := h.svc.GetConversation(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "conversation not found")
		return
	}
	util.JSON(c, http.StatusOK, conv)
}

func (h *Handler) UpdateConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	var body struct {
		CustomerID     *string `json:"customer_id"`
		AssignedUserID *string `json:"assigned_user_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	var customerID, assignedID *uuid.UUID
	if body.CustomerID != nil && *body.CustomerID != "" {
		if uid, err := uuid.Parse(*body.CustomerID); err == nil {
			customerID = &uid
		}
	}
	if body.AssignedUserID != nil && *body.AssignedUserID != "" {
		if uid, err := uuid.Parse(*body.AssignedUserID); err == nil {
			assignedID = &uid
		}
	}
	conv, err := h.svc.UpdateConversation(c.Request.Context(), id, customerID, assignedID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, conv)
}

func (h *Handler) ListAssignees(c *gin.Context) {
	items, err := h.svc.ListAssignees(c.Request.Context())
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func strVal(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
