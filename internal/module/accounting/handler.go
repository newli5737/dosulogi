package accounting

import (
	"net/http"
	"path/filepath"

	"github.com/dosu-logi/logistics-erp/internal/integration/sepay"
	"github.com/dosu-logi/logistics-erp/internal/middleware"
	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc        *Service
	sepaySecret string
}

func NewHandler(svc *Service, sepaySecret string) *Handler {
	return &Handler{svc: svc, sepaySecret: sepaySecret}
}

func (h *Handler) ListInvoices(c *gin.Context) {
	page, limit, offset := util.ParsePagination(c)
	items, total, err := h.svc.ListInvoices(c.Request.Context(), c.Query("status"), c.Query("customer_id"), c.Query("from"), c.Query("to"), limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) CreateInvoice(c *gin.Context) {
	var req struct {
		Invoice Invoice   `json:"invoice"`
		Items   []LineItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		var inv Invoice
		if err2 := c.ShouldBindJSON(&inv); err2 != nil {
			util.BadRequest(c, err.Error())
			return
		}
		req.Invoice = inv
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	req.Invoice.CreatedBy = &userID
	if err := h.svc.CreateInvoice(c.Request.Context(), &req.Invoice, req.Items); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, req.Invoice)
}

func (h *Handler) GetInvoice(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	inv, err := h.svc.GetInvoice(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "invoice not found")
		return
	}
	util.JSON(c, http.StatusOK, inv)
}

func (h *Handler) UpdateInvoice(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		Invoice Invoice    `json:"invoice"`
		Items   []LineItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	req.Invoice.ID = id
	if err := h.svc.UpdateInvoice(c.Request.Context(), &req.Invoice, req.Items); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, req.Invoice)
}

func (h *Handler) SendInvoice(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.SendInvoice(c.Request.Context(), id); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"message": "invoice sent"})
}

func (h *Handler) CancelInvoice(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.CancelInvoice(c.Request.Context(), id); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"message": "invoice cancelled"})
}

func (h *Handler) DownloadInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		util.BadRequest(c, "invalid invoice id")
		return
	}
	path, err := h.svc.DownloadInvoice(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "invoice not found" {
			util.NotFound(c, err.Error())
			return
		}
		util.InternalError(c, err.Error())
		return
	}
	c.FileAttachment(path, filepath.Base(path))
}

func (h *Handler) ListPayments(c *gin.Context) {
	page, limit, offset := util.ParsePagination(c)
	items, total, err := h.svc.ListPayments(c.Request.Context(), limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) CreatePayment(c *gin.Context) {
	var p Payment
	if err := c.ShouldBindJSON(&p); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.CreatePayment(c.Request.Context(), &p); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, p)
}

func (h *Handler) SePayWebhook(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if !sepay.VerifyAuth(auth, h.sepaySecret) {
		util.Unauthorized(c, "invalid authorization")
		return
	}
	payload, err := sepay.ParseWebhook(c.Request.Body)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.HandleSePayWebhook(c.Request.Context(), payload); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) RevenueReport(c *gin.Context) {
	items, err := h.svc.RevenueReport(c.Request.Context(), c.Query("from"), c.Query("to"), c.DefaultQuery("group_by", "month"))
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func (h *Handler) ARReport(c *gin.Context) {
	items, err := h.svc.ARReport(c.Request.Context())
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func (h *Handler) ListByCustomer(c *gin.Context) {
	customerID, _ := uuid.Parse(c.Param("id"))
	items, err := h.svc.ListByCustomer(c.Request.Context(), customerID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}
