package http

import (
	"errors"

	"github.com/dosu-logi/logistics-erp/internal/crm/application"
	"github.com/dosu-logi/logistics-erp/internal/crm/domain"
	"github.com/dosu-logi/logistics-erp/internal/middleware"
	"github.com/dosu-logi/logistics-erp/internal/platform/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CustomerHandler struct {
	svc *application.CustomerService
}

func NewCustomerHandler(svc *application.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) List(c *gin.Context) {
	page, limit, offset := httpx.ParsePageLimit(c)
	f := domain.CustomerFilter{
		Page: page, Limit: limit, Offset: offset,
		Query: c.Query("q"), Type: c.Query("type"), Segment: c.Query("segment"),
		Tier: c.Query("tier"), AssignedTo: c.Query("assigned_to"), Province: c.Query("province"),
		UserID: middleware.GetUserID(c), Role: middleware.GetRole(c),
	}
	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.List(c, items, httpx.Meta{Page: page, Limit: limit, Total: total})
}

func (h *CustomerHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", "invalid id")
		return
	}
	item, err := h.svc.GetDetail(c.Request.Context(), id)
	if err != nil {
		httpx.NotFound(c, "NOT_FOUND", "customer not found")
		return
	}
	httpx.OK(c, item)
}

func (h *CustomerHandler) ListContacts(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", "invalid id")
		return
	}
	items, err := h.svc.ListContacts(c.Request.Context(), customerID)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.OK(c, items)
}

func (h *CustomerHandler) CreateContact(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	var ct domain.Contact
	if err := c.ShouldBindJSON(&ct); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	ct.CustomerID = customerID
	item, err := h.svc.CreateContact(c.Request.Context(), &ct)
	if err != nil {
		mapErr(c, err)
		return
	}
	httpx.Created(c, item)
}

func (h *CustomerHandler) UpdateContact(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	contactID, _ := uuid.Parse(c.Param("contact_id"))
	var ct domain.Contact
	if err := c.ShouldBindJSON(&ct); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	ct.ID = contactID
	ct.CustomerID = customerID
	item, err := h.svc.UpdateContact(c.Request.Context(), &ct)
	if err != nil {
		mapErr(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *CustomerHandler) DeleteContact(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	contactID, _ := uuid.Parse(c.Param("contact_id"))
	if err := h.svc.DeleteContact(c.Request.Context(), customerID, contactID); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(204)
}

func (h *CustomerHandler) ListInteractions(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", "invalid id")
		return
	}
	page, limit, offset := httpx.ParsePageLimit(c)
	f := domain.InteractionFilter{
		Page: page, Limit: limit, Offset: offset,
		CustomerID: customerID, Channel: c.Query("channel"),
	}
	items, total, err := h.svc.ListInteractions(c.Request.Context(), f)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.List(c, items, httpx.Meta{Page: page, Limit: limit, Total: total})
}

func (h *CustomerHandler) CreateInteraction(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	var req application.CreateInteractionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	item, err := h.svc.CreateInteraction(c.Request.Context(), customerID, req, userID)
	if err != nil {
		mapErr(c, err)
		return
	}
	httpx.Created(c, item)
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var req application.CreateCustomerInput
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	req.CreatedBy = userID
	item, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		mapErr(c, err)
		return
	}
	httpx.Created(c, item)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var patch domain.Customer
	if err := c.ShouldBindJSON(&patch); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, patch)
	if err != nil {
		mapErr(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *CustomerHandler) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), id, middleware.GetRole(c)); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(204)
}

type TicketHandler struct {
	svc *application.TicketService
}

func NewTicketHandler(svc *application.TicketService) *TicketHandler {
	return &TicketHandler{svc: svc}
}

func (h *TicketHandler) List(c *gin.Context) {
	page, limit, offset := httpx.ParsePageLimit(c)
	f := domain.TicketFilter{
		Page: page, Limit: limit, Offset: offset,
		Status: c.Query("status"), Priority: c.Query("priority"),
		CustomerID: c.Query("customer_id"), Overdue: c.Query("overdue") == "true",
	}
	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.List(c, items, httpx.Meta{Page: page, Limit: limit, Total: total})
}

func (h *TicketHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	t, comments, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.NotFound(c, "NOT_FOUND", "ticket not found")
		return
	}
	httpx.OK(c, gin.H{"ticket": t, "comments": comments})
}

func (h *TicketHandler) Create(c *gin.Context) {
	var req application.CreateTicketInput
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	req.CreatedBy, _ = uuid.Parse(middleware.GetUserID(c))
	item, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		mapErr(c, err)
		return
	}
	httpx.Created(c, item)
}

func (h *TicketHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var patch domain.Ticket
	if err := c.ShouldBindJSON(&patch); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, patch)
	if err != nil {
		mapErr(c, err)
		return
	}
	httpx.OK(c, item)
}

func (h *TicketHandler) AddComment(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		Body       string `json:"body"`
		IsInternal bool   `json:"is_internal"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	if req.Body == "" {
		httpx.BadRequest(c, "VALIDATION_ERROR", "body required")
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	item, err := h.svc.AddComment(c.Request.Context(), id, req.Body, req.IsInternal, userID)
	if err != nil {
		mapErr(c, err)
		return
	}
	httpx.Created(c, item)
}

func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, application.ErrValidation):
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, application.ErrConflict):
		httpx.Conflict(c, "CONFLICT", "email already exists")
	case errors.Is(err, application.ErrPrimaryContact):
		httpx.BadRequest(c, "VALIDATION_ERROR", "cannot delete primary contact when it is the only contact")
	case errors.Is(err, application.ErrForbidden):
		httpx.Forbidden(c, "insufficient permissions")
	case errors.Is(err, application.ErrNotFound), errors.Is(err, pgx.ErrNoRows):
		httpx.NotFound(c, "NOT_FOUND", "resource not found")
	default:
		httpx.Internal(c, err.Error())
	}
}
