package sales

import (
	"fmt"
	"net/http"
	"path/filepath"

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

// Opportunities
func (h *Handler) ListOpportunities(c *gin.Context) {
	page, limit, offset := util.ParsePagination(c)
	items, total, err := h.svc.ListOpportunities(c.Request.Context(), c.Query("stage"), c.Query("assigned_to"),
		middleware.GetUserID(c), middleware.GetRole(c), limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) CreateOpportunity(c *gin.Context) {
	var o Opportunity
	if err := c.ShouldBindJSON(&o); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	o.CreatedBy = &userID
	if err := h.svc.CreateOpportunity(c.Request.Context(), &o); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, o)
}

func (h *Handler) GetOpportunity(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	o, err := h.svc.GetOpportunity(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "opportunity not found")
		return
	}
	util.JSON(c, http.StatusOK, o)
}

func (h *Handler) UpdateOpportunity(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var o Opportunity
	if err := c.ShouldBindJSON(&o); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	o.ID = id
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	if err := h.svc.UpdateOpportunity(c.Request.Context(), &o, userID); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, o)
}

func (h *Handler) DeleteOpportunity(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.DeleteOpportunity(c.Request.Context(), id); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// Contracts
func (h *Handler) ListContracts(c *gin.Context) {
	page, limit, offset := util.ParsePagination(c)
	items, total, err := h.svc.ListContracts(c.Request.Context(), c.Query("status"), c.Query("customer_id"), limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) CreateContract(c *gin.Context) {
	var ct Contract
	if err := c.ShouldBindJSON(&ct); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	ct.CreatedBy = &userID
	if err := h.svc.CreateContract(c.Request.Context(), &ct); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, ct)
}

func (h *Handler) GetContract(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	ct, err := h.svc.GetContract(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "contract not found")
		return
	}
	util.JSON(c, http.StatusOK, ct)
}

func (h *Handler) UpdateContract(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var ct Contract
	if err := c.ShouldBindJSON(&ct); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	ct.ID = id
	if err := h.svc.UpdateContract(c.Request.Context(), &ct); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, ct)
}

func (h *Handler) UploadContract(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	file, err := c.FormFile("file")
	if err != nil {
		util.BadRequest(c, "file required")
		return
	}
	dest := filepath.Join(h.svc.uploadDir, "contracts", fmt.Sprintf("%s.pdf", id))
	if err := c.SaveUploadedFile(file, dest); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	ct, err := h.svc.UploadContractFile(c.Request.Context(), id, dest)
	if err != nil {
		util.NotFound(c, "contract not found")
		return
	}
	util.JSON(c, http.StatusOK, ct)
}

// Quotations
func (h *Handler) ListQuotations(c *gin.Context) {
	page, limit, offset := util.ParsePagination(c)
	items, total, err := h.svc.ListQuotations(c.Request.Context(), c.Query("status"), c.Query("customer_id"), limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) CreateQuotation(c *gin.Context) {
	var qt Quotation
	if err := c.ShouldBindJSON(&qt); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	qt.CreatedBy = &userID
	if err := h.svc.CreateQuotation(c.Request.Context(), &qt); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, qt)
}

func (h *Handler) GetQuotation(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	qt, err := h.svc.GetQuotation(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "quotation not found")
		return
	}
	util.JSON(c, http.StatusOK, qt)
}

func (h *Handler) UpdateQuotation(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var qt Quotation
	if err := c.ShouldBindJSON(&qt); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	qt.ID = id
	if err := h.svc.UpdateQuotation(c.Request.Context(), &qt); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, qt)
}

func (h *Handler) SendQuotation(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.SendQuotation(c.Request.Context(), id); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"message": "quotation sent"})
}

func (h *Handler) ConvertQuotation(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	ct, err := h.svc.ConvertQuotation(c.Request.Context(), id, userID)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, ct)
}
