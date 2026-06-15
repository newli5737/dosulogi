package crm

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

func (h *Handler) List(c *gin.Context) {
	page, limit, offset := util.ParsePagination(c)
	f := CustomerFilter{
		Segment: c.Query("segment"), Tier: c.Query("tier"),
		AssignedTo: c.Query("assigned_to"), Query: c.Query("q"),
		UserID: middleware.GetUserID(c), Role: middleware.GetRole(c),
		Limit: limit, Offset: offset,
	}
	items, total, err := h.svc.ListCustomers(c.Request.Context(), f)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	item, err := h.svc.CreateCustomer(c.Request.Context(), req, userID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, item)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	item, err := h.svc.GetCustomer(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "customer not found")
		return
	}
	util.JSON(c, http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.UpdateCustomer(c.Request.Context(), id, req, middleware.GetUserID(c), middleware.GetRole(c))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			util.NotFound(c, "customer not found")
			return
		}
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, item)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.DeleteCustomer(c.Request.Context(), id); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListContacts(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	items, err := h.svc.ListContacts(c.Request.Context(), customerID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, items)
}

func (h *Handler) CreateContact(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.CreateContact(c.Request.Context(), customerID, req)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, item)
}

func (h *Handler) UpdateContact(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	contactID, _ := uuid.Parse(c.Param("contact_id"))
	var req CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	item, err := h.svc.UpdateContact(c.Request.Context(), customerID, contactID, req)
	if err != nil {
		util.NotFound(c, "contact not found")
		return
	}
	util.JSON(c, http.StatusOK, item)
}

func (h *Handler) DeleteContact(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	contactID, _ := uuid.Parse(c.Param("contact_id"))
	if err := h.svc.DeleteContact(c.Request.Context(), customerID, contactID); err != nil {
		util.NotFound(c, "contact not found")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListInteractions(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	items, err := h.svc.ListInteractions(c.Request.Context(), customerID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, items)
}

func (h *Handler) CreateInteraction(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	var req CreateInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	item, err := h.svc.CreateInteraction(c.Request.Context(), customerID, req, userID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, item)
}
