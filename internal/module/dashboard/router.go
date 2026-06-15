package dashboard

import (
	"net/http"

	"github.com/dosu-logi/logistics-erp/internal/module/tracking"
	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/gin-gonic/gin"
)

type Router struct {
	handler        *Handler
	trackingSvc    *tracking.Service
}

func NewRouter(dbHandler *Handler, trackingSvc *tracking.Service) *Router {
	return &Router{handler: dbHandler, trackingSvc: trackingSvc}
}

func (r *Router) Summary(c *gin.Context) {
	s, err := r.handler.summary(c.Request.Context())
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSON(c, http.StatusOK, s)
}

func (r *Router) SalesFunnel(c *gin.Context) {
	items, err := r.handler.salesFunnel(c.Request.Context())
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func (r *Router) RevenueTrend(c *gin.Context) {
	items, err := r.handler.revenueTrend(c.Request.Context())
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func (r *Router) TicketStats(c *gin.Context) {
	items, err := r.handler.ticketStats(c.Request.Context())
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func (r *Router) ShipmentStats(c *gin.Context) {
	items, err := r.handler.shipmentStats(c.Request.Context())
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}

func (r *Router) ShipmentMap(c *gin.Context) {
	items, err := r.trackingSvc.ListMapPoints(c.Request.Context())
	if err != nil {
		util.InternalError(c, err.Error())
		return
	}
	util.JSONList(c, http.StatusOK, items)
}
