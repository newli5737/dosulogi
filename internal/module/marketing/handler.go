package marketing

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

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
	items, total, err := h.svc.List(c.Request.Context(), c.Query("status"), limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) Create(c *gin.Context) {
	var camp Campaign
	if err := c.ShouldBindJSON(&camp); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	camp.CreatedBy = &userID
	if err := h.svc.Create(c.Request.Context(), &camp); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, camp)
}

func (h *Handler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	camp, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "campaign not found")
		return
	}
	util.JSON(c, http.StatusOK, camp)
}

func (h *Handler) Update(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var camp Campaign
	if err := c.ShouldBindJSON(&camp); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	camp.ID = id
	if err := h.svc.Update(c.Request.Context(), &camp); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, camp)
}

func (h *Handler) Send(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.Send(c.Request.Context(), id); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"message": "campaign sent"})
}

func (h *Handler) Schedule(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var req ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	camp, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "campaign not found")
		return
	}
	camp.ScheduledAt = &req.ScheduledAt
	camp.Status = "scheduled"
	if err := h.svc.Update(c.Request.Context(), camp); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, camp)
}

func (h *Handler) ListLogs(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	page, limit, offset := util.ParsePagination(c)
	items, total, err := h.svc.ListLogs(c.Request.Context(), id, c.Query("status"), limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) EmailWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		util.BadRequest(c, "invalid body")
		return
	}
	var events []EmailWebhookEvent
	if err := json.Unmarshal(body, &events); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.HandleEmailWebhook(c.Request.Context(), events); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"ok": true})
}

// Scheduler runs scheduled campaigns
type Scheduler struct {
	svc      *Service
	interval time.Duration
	stop     chan struct{}
}

func NewScheduler(svc *Service) *Scheduler {
	return &Scheduler{svc: svc, interval: time.Minute, stop: make(chan struct{})}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = s.svc.ProcessScheduled(ctx)
			case <-s.stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() { close(s.stop) }
