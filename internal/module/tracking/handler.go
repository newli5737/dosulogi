package tracking

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc           *Service
	webhookSecret string
}

func NewHandler(svc *Service, webhookSecret string) *Handler {
	return &Handler{svc: svc, webhookSecret: webhookSecret}
}

func (h *Handler) Create(c *gin.Context) {
	var sh Shipment
	if err := c.ShouldBindJSON(&sh); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	if sh.TrackingCode == "" {
		util.BadRequest(c, "tracking_code required")
		return
	}
	if err := h.svc.Create(c.Request.Context(), &sh); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusCreated, sh)
}

func (h *Handler) List(c *gin.Context) {
	page, limit, offset := util.ParsePagination(c)
	items, total, err := h.svc.List(c.Request.Context(), c.Query("status"), c.Query("customer_id"), limit, offset)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.Paginated(c, items, page, limit, total)
}

func (h *Handler) Get(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		util.NotFound(c, "shipment not found")
		return
	}
	util.JSON(c, http.StatusOK, item)
}

func (h *Handler) ListEvents(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	items, err := h.svc.ListEvents(c.Request.Context(), id)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func (h *Handler) Sync(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	item, err := h.svc.SyncShipment(c.Request.Context(), id)
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, item)
}

func (h *Handler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		util.BadRequest(c, "invalid body")
		return
	}
	sig := c.GetHeader("X-Hmac-Signature")
	if !verifyHMAC(body, sig, h.webhookSecret) {
		util.Unauthorized(c, "invalid signature")
		return
	}
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.HandleWebhook(c.Request.Context(), payload); err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, gin.H{"ok": true})
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

func verifyHMAC(body []byte, sig, secret string) bool {
	if secret == "" || sig == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

type Poller struct {
	svc      *Service
	interval time.Duration
	stop     chan struct{}
}

func NewPoller(svc *Service, interval time.Duration) *Poller {
	return &Poller{svc: svc, interval: interval, stop: make(chan struct{})}
}

func (p *Poller) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = p.svc.PollActive(ctx)
			case <-p.stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *Poller) Stop() { close(p.stop) }
