package router

import (
	"github.com/dosu-logi/logistics-erp/internal/config"
	"github.com/dosu-logi/logistics-erp/internal/integration/mailer"
	"github.com/dosu-logi/logistics-erp/internal/integration/tracking3p"
	"github.com/dosu-logi/logistics-erp/internal/middleware"
	"github.com/dosu-logi/logistics-erp/internal/module/accounting"
	"github.com/dosu-logi/logistics-erp/internal/module/auth"
	"github.com/dosu-logi/logistics-erp/internal/module/crm"
	"github.com/dosu-logi/logistics-erp/internal/module/dashboard"
	"github.com/dosu-logi/logistics-erp/internal/module/marketing"
	"github.com/dosu-logi/logistics-erp/internal/module/sales"
	"github.com/dosu-logi/logistics-erp/internal/module/tracking"
	"github.com/dosu-logi/logistics-erp/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Config  *config.Config
	DB      *pgxpool.Pool
	JWT     *util.JWTManager
	Poller  *tracking.Poller
	Sched   *marketing.Scheduler
}

func Setup(deps Deps) *gin.Engine {
	if deps.Config.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.CORS(deps.Config.CORSOrigins))

	mailer := mailer.New(deps.Config.FromEmail)
	trackClient := tracking3p.NewClient(deps.Config.TrackingAPIBaseURL, deps.Config.TrackingAPIKey)

	authRepo := auth.NewRepository(deps.DB)
	authSvc := auth.NewService(authRepo, deps.JWT)
	authH := auth.NewHandler(authSvc)

	crmRepo := crm.NewRepository(deps.DB)
	crmSvc := crm.NewService(crmRepo)
	crmH := crm.NewHandler(crmSvc)

	salesRepo := sales.NewRepository(deps.DB)
	salesSvc := sales.NewService(salesRepo, mailer, deps.Config.UploadDir)
	salesH := sales.NewHandler(salesSvc)

	trackRepo := tracking.NewRepository(deps.DB)
	trackSvc := tracking.NewService(trackRepo, trackClient)
	trackH := tracking.NewHandler(trackSvc, deps.Config.TrackingWebhookSecret)

	acctRepo := accounting.NewRepository(deps.DB)
	acctSvc := accounting.NewService(acctRepo, mailer, deps.Config.UploadDir)
	acctH := accounting.NewHandler(acctSvc, deps.Config.SePayWebhookSecret)

	mktRepo := marketing.NewRepository(deps.DB)
	mktSvc := marketing.NewService(mktRepo, mailer)
	mktH := marketing.NewHandler(mktSvc)

	dashH := dashboard.NewRouter(dashboard.NewHandler(deps.DB), trackSvc)

	api := r.Group("/api/v1")

	// Public auth
	api.POST("/auth/login", authH.Login)
	api.POST("/auth/refresh", authH.Refresh)

	// Public webhooks
	api.POST("/webhooks/tracking", trackH.Webhook)
	api.POST("/webhooks/sepay", acctH.SePayWebhook)
	api.POST("/webhooks/email", mktH.EmailWebhook)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth(deps.JWT))
	protected.Use(middleware.RBAC())
	protected.Use(middleware.DirectorReadOnly())

	// Auth
	protected.POST("/auth/logout", authH.Logout)
	protected.GET("/auth/me", authH.Me)
	protected.PUT("/auth/me/password", authH.ChangePassword)

	// Users (admin)
	users := protected.Group("/users")
	users.Use(middleware.RequireRoles("admin"))
	users.GET("", authH.ListUsers)
	users.POST("", authH.CreateUser)
	users.GET("/:id", authH.GetUser)
	users.PUT("/:id", authH.UpdateUser)
	users.DELETE("/:id", authH.DeleteUser)

	// CRM
	protected.GET("/customers", crmH.List)
	protected.POST("/customers", crmH.Create)
	protected.GET("/customers/:id", crmH.Get)
	protected.PUT("/customers/:id", crmH.Update)
	protected.DELETE("/customers/:id", crmH.Delete)
	protected.GET("/customers/:id/contacts", crmH.ListContacts)
	protected.POST("/customers/:id/contacts", crmH.CreateContact)
	protected.PUT("/customers/:id/contacts/:contact_id", crmH.UpdateContact)
	protected.DELETE("/customers/:id/contacts/:contact_id", crmH.DeleteContact)
	protected.GET("/customers/:id/interactions", crmH.ListInteractions)
	protected.POST("/customers/:id/interactions", crmH.CreateInteraction)
	protected.GET("/customers/:id/shipments", trackH.ListByCustomer)
	protected.GET("/customers/:id/invoices", acctH.ListByCustomer)
	protected.GET("/customers/:id/contracts", listContractsByCustomer(salesSvc))

	// Sales
	protected.GET("/opportunities", salesH.ListOpportunities)
	protected.POST("/opportunities", salesH.CreateOpportunity)
	protected.GET("/opportunities/:id", salesH.GetOpportunity)
	protected.PUT("/opportunities/:id", salesH.UpdateOpportunity)
	protected.DELETE("/opportunities/:id", salesH.DeleteOpportunity)

	protected.GET("/contracts", salesH.ListContracts)
	protected.POST("/contracts", salesH.CreateContract)
	protected.GET("/contracts/:id", salesH.GetContract)
	protected.PUT("/contracts/:id", salesH.UpdateContract)
	protected.POST("/contracts/:id/upload", salesH.UploadContract)

	protected.GET("/quotations", salesH.ListQuotations)
	protected.POST("/quotations", salesH.CreateQuotation)
	protected.GET("/quotations/:id", salesH.GetQuotation)
	protected.PUT("/quotations/:id", salesH.UpdateQuotation)
	protected.POST("/quotations/:id/send", salesH.SendQuotation)
	protected.POST("/quotations/:id/convert", salesH.ConvertQuotation)

	// Tracking
	protected.GET("/shipments", trackH.List)
	protected.POST("/shipments", trackH.Create)
	protected.GET("/shipments/:id", trackH.Get)
	protected.GET("/shipments/:id/events", trackH.ListEvents)
	protected.POST("/shipments/:id/sync", trackH.Sync)

	// Accounting
	protected.GET("/invoices", acctH.ListInvoices)
	protected.POST("/invoices", acctH.CreateInvoice)
	protected.GET("/invoices/:id", acctH.GetInvoice)
	protected.PUT("/invoices/:id", acctH.UpdateInvoice)
	protected.POST("/invoices/:id/send", acctH.SendInvoice)
	protected.POST("/invoices/:id/cancel", acctH.CancelInvoice)
	protected.GET("/invoices/:id/download", acctH.DownloadInvoice)
	protected.GET("/payments", acctH.ListPayments)
	protected.POST("/payments", acctH.CreatePayment)
	protected.GET("/reports/revenue", acctH.RevenueReport)
	protected.GET("/reports/ar", acctH.ARReport)

	// Marketing
	protected.GET("/campaigns", mktH.List)
	protected.POST("/campaigns", mktH.Create)
	protected.GET("/campaigns/:id", mktH.Get)
	protected.PUT("/campaigns/:id", mktH.Update)
	protected.POST("/campaigns/:id/send", mktH.Send)
	protected.POST("/campaigns/:id/schedule", mktH.Schedule)
	protected.GET("/campaigns/:id/logs", mktH.ListLogs)

	// Dashboard
	protected.GET("/dashboard/summary", dashH.Summary)
	protected.GET("/dashboard/sales-funnel", dashH.SalesFunnel)
	protected.GET("/dashboard/shipment-map", dashH.ShipmentMap)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

func listContractsByCustomer(svc *sales.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("id")
		page, limit, offset := util.ParsePagination(c)
		items, total, err := svc.ListContracts(c.Request.Context(), "", customerID, limit, offset)
		if err != nil {
			util.InternalError(c, err.Error())
			return
		}
		util.Paginated(c, items, page, limit, total)
	}
}
