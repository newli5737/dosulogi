package router

import (
	crmapp "github.com/dosu-logi/logistics-erp/internal/crm/application"
	crmhttp "github.com/dosu-logi/logistics-erp/internal/crm/adapter/http"
	crmrepo "github.com/dosu-logi/logistics-erp/internal/crm/adapter/postgres"
	salesapp "github.com/dosu-logi/logistics-erp/internal/sales/application"
	saleshttp "github.com/dosu-logi/logistics-erp/internal/sales/adapter/http"
	salesmailer "github.com/dosu-logi/logistics-erp/internal/sales/adapter/mailer"
	salesrepo "github.com/dosu-logi/logistics-erp/internal/sales/adapter/postgres"
	"github.com/dosu-logi/logistics-erp/internal/config"
	"github.com/dosu-logi/logistics-erp/internal/integration/mailer"
	"github.com/dosu-logi/logistics-erp/internal/integration/tracking3p"
	"github.com/dosu-logi/logistics-erp/internal/middleware"
	"github.com/dosu-logi/logistics-erp/internal/module/accounting"
	"github.com/dosu-logi/logistics-erp/internal/module/auth"
	"github.com/dosu-logi/logistics-erp/internal/module/dashboard"
	"github.com/dosu-logi/logistics-erp/internal/module/marketing"
	"github.com/dosu-logi/logistics-erp/internal/chat"
	"github.com/dosu-logi/logistics-erp/internal/chat/adapter/zalo"
	"github.com/dosu-logi/logistics-erp/internal/module/tracking"
	"github.com/dosu-logi/logistics-erp/internal/platform/cache"
	"github.com/dosu-logi/logistics-erp/internal/platform/httpx"
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
	Cache   *cache.Store
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
	authSvc := auth.NewService(authRepo, deps.JWT, deps.Config.JWTAdminRefreshTTL)
	authH := auth.NewHandler(authSvc, deps.Config.CookieDomain)

	custRepo := crmrepo.NewCustomerRepo(deps.DB)
	contactRepo := crmrepo.NewContactRepo(deps.DB)
	interactionRepo := crmrepo.NewInteractionRepo(deps.DB)
	ticketRepo := crmrepo.NewTicketRepo(deps.DB)
	ticketCommentRepo := crmrepo.NewTicketCommentRepo(deps.DB)
	custSvc := crmapp.NewCustomerService(custRepo, contactRepo, interactionRepo)
	ticketSvc := crmapp.NewTicketService(ticketRepo, ticketCommentRepo)
	custHex := crmhttp.NewCustomerHandler(custSvc)
	ticketHex := crmhttp.NewTicketHandler(ticketSvc)

	oppRepo := salesrepo.NewOpportunityRepo(deps.DB)
	contractRepo := salesrepo.NewContractRepo(deps.DB)
	quoteRepo := salesrepo.NewQuotationRepo(deps.DB)
	mailAdapter := salesmailer.New(mailer)
	oppSvc := salesapp.NewOpportunityService(oppRepo)
	contractSvc := salesapp.NewContractService(contractRepo)
	quoteSvc := salesapp.NewQuotationService(quoteRepo, contractRepo, mailAdapter)
	oppHex := saleshttp.NewOpportunityHandler(oppSvc)
	contractHex := saleshttp.NewContractHandler(contractSvc, deps.Config.UploadDir)
	quoteHex := saleshttp.NewQuotationHandler(quoteSvc)

	trackRepo := tracking.NewRepository(deps.DB)
	trackSvc := tracking.NewService(trackRepo, trackClient)
	trackH := tracking.NewHandler(trackSvc, deps.Config.TrackingWebhookSecret)

	acctRepo := accounting.NewRepository(deps.DB)
	acctSvc := accounting.NewService(acctRepo, mailer, deps.Config.UploadDir, accounting.CompanyInfo{
		Name:    deps.Config.CompanyName,
		TaxCode: deps.Config.CompanyTaxCode,
		Address: deps.Config.CompanyAddress,
		Phone:   deps.Config.CompanyPhone,
		Email:   deps.Config.CompanyEmail,
		Tagline: deps.Config.CompanyTagline,
	})
	acctH := accounting.NewHandler(acctSvc, deps.Config.SePayWebhookSecret)

	mktRepo := marketing.NewRepository(deps.DB)
	mktSvc := marketing.NewService(mktRepo, mailer)
	mktH := marketing.NewHandler(mktSvc)

	dashH := dashboard.NewRouter(dashboard.NewHandler(deps.DB), trackSvc)

	chatRepo := chat.NewRepository(deps.DB)
	chatZalo := zalo.NewClient(deps.Config.ZaloBridgeURL)
	chatSvc := chat.NewService(chatRepo, chatZalo)
	chatH := chat.NewHandler(chatSvc)

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
	protected.Use(middleware.RateLimit(deps.Cache, 120))
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

	// CRM (hexagonal customers + tickets)
	protected.GET("/customers", custHex.List)
	protected.POST("/customers", custHex.Create)
	protected.GET("/customers/:id", custHex.Get)
	protected.PUT("/customers/:id", custHex.Update)
	protected.DELETE("/customers/:id", custHex.Delete)
	protected.GET("/tickets", ticketHex.List)
	protected.POST("/tickets", ticketHex.Create)
	protected.GET("/tickets/:id", ticketHex.Get)
	protected.PUT("/tickets/:id", ticketHex.Update)
	protected.POST("/tickets/:id/comments", ticketHex.AddComment)
	protected.GET("/customers/:id/contacts", custHex.ListContacts)
	protected.POST("/customers/:id/contacts", custHex.CreateContact)
	protected.PUT("/customers/:id/contacts/:contact_id", custHex.UpdateContact)
	protected.DELETE("/customers/:id/contacts/:contact_id", custHex.DeleteContact)
	protected.GET("/customers/:id/interactions", custHex.ListInteractions)
	protected.POST("/customers/:id/interactions", custHex.CreateInteraction)
	protected.GET("/customers/:id/shipments", trackH.ListByCustomer)
	protected.GET("/customers/:id/invoices", acctH.ListByCustomer)
	protected.GET("/customers/:id/contracts", listContractsByCustomer(contractSvc))

	// Sales (hexagonal)
	protected.GET("/opportunities", oppHex.List)
	protected.POST("/opportunities", oppHex.Create)
	protected.GET("/opportunities/:id", oppHex.Get)
	protected.PUT("/opportunities/:id", oppHex.Update)
	protected.DELETE("/opportunities/:id", oppHex.Delete)
	protected.GET("/opportunities/:id/stage-history", oppHex.StageHistory)

	protected.GET("/contracts", contractHex.List)
	protected.POST("/contracts", contractHex.Create)
	protected.GET("/contracts/:id", contractHex.Get)
	protected.PUT("/contracts/:id", contractHex.Update)
	protected.POST("/contracts/:id/upload", contractHex.Upload)

	protected.GET("/quotations", quoteHex.List)
	protected.POST("/quotations", quoteHex.Create)
	protected.GET("/quotations/:id", quoteHex.Get)
	protected.PUT("/quotations/:id", quoteHex.Update)
	protected.POST("/quotations/:id/send", quoteHex.Send)
	protected.POST("/quotations/:id/convert", quoteHex.Convert)

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
	protected.GET("/dashboard/revenue-trend", dashH.RevenueTrend)
	protected.GET("/dashboard/ticket-stats", dashH.TicketStats)
	protected.GET("/dashboard/shipment-stats", dashH.ShipmentStats)
	protected.GET("/dashboard/shipment-map", dashH.ShipmentMap)

	// Omnichannel chat
	protected.GET("/chat/accounts", chatH.ListAccounts)
	protected.POST("/chat/accounts", chatH.CreateAccount)
	protected.PUT("/chat/accounts/:id", chatH.UpdateAccount)
	protected.DELETE("/chat/accounts/:id", chatH.DeleteAccount)
	protected.POST("/chat/accounts/:id/zalo/qr", chatH.ZaloQRLogin)
	protected.GET("/chat/accounts/:id/zalo/status", chatH.ZaloLoginStatus)
	protected.GET("/chat/inbox", chatH.Inbox)
	protected.GET("/chat/threads/:threadId", chatH.Thread)
	protected.POST("/chat/send", chatH.Send)
	protected.GET("/chat/conversations/:id", chatH.GetConversation)
	protected.PUT("/chat/conversations/:id", chatH.UpdateConversation)
	protected.GET("/chat/assignees", chatH.ListAssignees)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

func listContractsByCustomer(svc *salesapp.ContractService) gin.HandlerFunc {
	return func(c *gin.Context) {
		customerID := c.Param("id")
		page, limit, offset := httpx.ParsePageLimit(c)
		items, total, err := svc.List(c.Request.Context(), "", customerID, limit, offset)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		httpx.List(c, items, httpx.Meta{Page: page, Limit: limit, Total: total})
	}
}
