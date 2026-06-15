package http

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/dosu-logi/logistics-erp/internal/middleware"
	"github.com/dosu-logi/logistics-erp/internal/platform/httpx"
	"github.com/dosu-logi/logistics-erp/internal/sales/application"
	"github.com/dosu-logi/logistics-erp/internal/sales/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OpportunityHandler struct{ svc *application.OpportunityService }

func NewOpportunityHandler(svc *application.OpportunityService) *OpportunityHandler {
	return &OpportunityHandler{svc: svc}
}

func (h *OpportunityHandler) List(c *gin.Context) {
	page, limit, offset := httpx.ParsePageLimit(c)
	f := domain.OpportunityFilter{
		Page: page, Limit: limit, Offset: offset,
		Stage: c.Query("stage"), AssignedTo: c.Query("assigned_to"),
		UserID: middleware.GetUserID(c), Role: middleware.GetRole(c),
	}
	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.List(c, items, httpx.Meta{Page: page, Limit: limit, Total: total})
}

func (h *OpportunityHandler) Create(c *gin.Context) {
	var req application.CreateOpportunityInput
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	req.CreatedBy = &userID
	if err := h.svc.Create(c.Request.Context(), &req); err != nil {
		mapErr(c, err)
		return
	}
	httpx.Created(c, req.Opportunity)
}

func (h *OpportunityHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	o, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.NotFound(c, "NOT_FOUND", "opportunity not found")
		return
	}
	httpx.OK(c, o)
}

func (h *OpportunityHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req struct {
		domain.Opportunity
		ShipmentIDs []uuid.UUID `json:"shipment_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	req.ID = id
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	if err := h.svc.Update(c.Request.Context(), &req.Opportunity, req.ShipmentIDs, userID); err != nil {
		mapErr(c, err)
		return
	}
	httpx.OK(c, req.Opportunity)
}

func (h *OpportunityHandler) StageHistory(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	items, err := h.svc.ListStageHistory(c.Request.Context(), id)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.OK(c, items)
}

func (h *OpportunityHandler) Delete(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(204)
}

type QuotationHandler struct{ svc *application.QuotationService }

func NewQuotationHandler(svc *application.QuotationService) *QuotationHandler {
	return &QuotationHandler{svc: svc}
}

func (h *QuotationHandler) List(c *gin.Context) {
	page, limit, offset := httpx.ParsePageLimit(c)
	items, total, err := h.svc.List(c.Request.Context(), c.Query("status"), c.Query("customer_id"), limit, offset)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.List(c, items, httpx.Meta{Page: page, Limit: limit, Total: total})
}

func (h *QuotationHandler) Create(c *gin.Context) {
	var qt domain.Quotation
	if err := c.ShouldBindJSON(&qt); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	qt.CreatedBy = &userID
	if err := h.svc.Create(c.Request.Context(), &qt); err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.Created(c, qt)
}

func (h *QuotationHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	qt, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.NotFound(c, "NOT_FOUND", "quotation not found")
		return
	}
	httpx.OK(c, qt)
}

func (h *QuotationHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var qt domain.Quotation
	if err := c.ShouldBindJSON(&qt); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	qt.ID = id
	if err := h.svc.Update(c.Request.Context(), &qt); err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.OK(c, qt)
}

func (h *QuotationHandler) Send(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.Send(c.Request.Context(), id); err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.OK(c, gin.H{"message": "quotation sent"})
}

func (h *QuotationHandler) Convert(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	ct, err := h.svc.Convert(c.Request.Context(), id, userID)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.Created(c, ct)
}

type ContractHandler struct {
	svc       *application.ContractService
	uploadDir string
}

func NewContractHandler(svc *application.ContractService, uploadDir string) *ContractHandler {
	return &ContractHandler{svc: svc, uploadDir: uploadDir}
}

func (h *ContractHandler) List(c *gin.Context) {
	page, limit, offset := httpx.ParsePageLimit(c)
	items, total, err := h.svc.List(c.Request.Context(), c.Query("status"), c.Query("customer_id"), limit, offset)
	if err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.List(c, items, httpx.Meta{Page: page, Limit: limit, Total: total})
}

func (h *ContractHandler) Create(c *gin.Context) {
	var ct domain.Contract
	if err := c.ShouldBindJSON(&ct); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	ct.CreatedBy = &userID
	if err := h.svc.Create(c.Request.Context(), &ct); err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.Created(c, ct)
}

func (h *ContractHandler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	ct, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.NotFound(c, "NOT_FOUND", "contract not found")
		return
	}
	httpx.OK(c, ct)
}

func (h *ContractHandler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var ct domain.Contract
	if err := c.ShouldBindJSON(&ct); err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
		return
	}
	ct.ID = id
	if err := h.svc.Update(c.Request.Context(), &ct); err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	httpx.OK(c, ct)
}

func (h *ContractHandler) Upload(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	file, err := c.FormFile("file")
	if err != nil {
		httpx.BadRequest(c, "VALIDATION_ERROR", "file required")
		return
	}
	dest := filepath.Join(h.uploadDir, "contracts", fmt.Sprintf("%s.pdf", id))
	if err := c.SaveUploadedFile(file, dest); err != nil {
		httpx.Internal(c, err.Error())
		return
	}
	ct, err := h.svc.UploadFile(c.Request.Context(), id, dest)
	if err != nil {
		httpx.NotFound(c, "NOT_FOUND", "contract not found")
		return
	}
	httpx.OK(c, ct)
}

func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, application.ErrValidation):
		httpx.BadRequest(c, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, pgx.ErrNoRows):
		httpx.NotFound(c, "NOT_FOUND", "resource not found")
	default:
		httpx.Internal(c, err.Error())
	}
}
