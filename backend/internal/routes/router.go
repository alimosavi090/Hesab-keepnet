package routes

import (
	"log/slog"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/auth"
	"github.com/ali/hesab-keepnet/backend/internal/config"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/handlers"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/middleware"
	"github.com/ali/hesab-keepnet/backend/internal/requestid"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/ali/hesab-keepnet/backend/internal/version"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, db *database.DB, log *slog.Logger) *gin.Engine {
	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	httpx.SetProduction(cfg.IsProd())

	engine := gin.New()
	if err := engine.SetTrustedProxies(nil); err != nil {
		log.Warn("set trusted proxies failed", "err", err)
	}

	engine.Use(
		requestid.Middleware(),
		middleware.Recovery(cfg.AppEnv, log),
		middleware.Logger(log),
		middleware.SecureHeaders(),
		middleware.CORS(cfg.CorsOrigins),
	)

	engine.NoRoute(func(c *gin.Context) {
		httpx.HandleError(c, httpx.ErrRouteNotFound())
	})
	engine.HandleMethodNotAllowed = true
	engine.NoMethod(func(c *gin.Context) {
		httpx.HandleError(c, httpx.ErrMethodNotAllowed())
	})

	svcs := services.NewServices(db.DB)
	authManager := auth.NewManager(db.DB)

	health := handlers.NewHealthHandler(db, cfg.AppEnv, version.Version)
	engine.GET("/health", health.Status)

	authHandler := handlers.NewAuthHandler(authManager, cfg.CookieSecure)
	financial := handlers.NewFinancialHandlers(svcs, services.NewBackupService(db.DB, cfg.BackupDir))

	apiV1 := engine.Group("/api/v1")
	apiV1.Use(middleware.RateLimit(300, time.Minute))

	apiV1.POST("/auth/login", middleware.RateLimit(5, time.Minute), authHandler.Login)
	apiV1.GET("/auth/csrf", middleware.RateLimit(30, time.Minute), authHandler.CSRF)

	protected := apiV1.Group("", middleware.Authenticate(authManager), middleware.RequireCSRF())
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/me", authHandler.Me)

	protected.GET("/bank-accounts", financial.ListBankAccounts)
	protected.POST("/bank-accounts", financial.CreateBankAccount)
	protected.GET("/bank-accounts/:id/balance", financial.GetBankAccountBalance)
	protected.PATCH("/bank-accounts/:id", financial.UpdateBankAccountActive)
	protected.PATCH("/bank-accounts/:id/edit", financial.UpdateBankAccount)
	protected.PATCH("/sales/:id", financial.UpdateSale)
	protected.PATCH("/expenses/:id", financial.UpdateExpense)

	protected.GET("/categories", financial.ListCategories)
	protected.POST("/categories", financial.CreateCategory)
	protected.DELETE("/categories/:id", financial.DeactivateCategory)

	protected.GET("/expenses", financial.ListExpenses)
	protected.POST("/expenses", financial.CreateExpense)
	protected.DELETE("/expenses/:id", financial.DeleteExpense)

	protected.GET("/sales", financial.ListSales)
	protected.POST("/sales", financial.CreateSale)
	protected.GET("/sales/:id", financial.GetSale)
	protected.DELETE("/sales/:id", financial.DeleteSale)

	protected.GET("/transfers", financial.ListTransfers)
	protected.POST("/transfers", financial.CreateTransfer)
	protected.DELETE("/transfers/:id", financial.DeleteTransfer)

	protected.GET("/representatives", financial.ListRepresentatives)
	protected.POST("/representatives", financial.CreateRepresentative)
	protected.GET("/representatives/:id/balance", financial.GetRepresentativeBalance)
	protected.GET("/representatives/:id/transactions", financial.ListRepresentativeTransactions)
	protected.POST("/representatives/:id/transactions", financial.CreateRepresentativeTransaction)
	protected.DELETE("/rep-transactions/:id", financial.DeleteRepresentativeTransaction)

	protected.GET("/reminders", financial.ListReminders)
	protected.GET("/reminders/upcoming", financial.UpcomingReminders)
	protected.POST("/reminders", financial.CreateReminder)
	protected.PATCH("/reminders/:id", financial.UpdateReminderDone)
	protected.DELETE("/reminders/:id", financial.DeleteReminder)

	protected.GET("/transactions", financial.LedgerFeed)
	protected.GET("/reports/summary", financial.ReportOverview)
	protected.GET("/reports/export.csv", financial.ExportCSV)

	// Notes: data-attached + daily journal
	protected.GET("/notes", financial.ListNotes)
	protected.POST("/notes", financial.CreateNote)
	protected.PATCH("/notes/:id", financial.UpdateNote)
	protected.DELETE("/notes/:id", financial.DeleteNote)

	// Backups
	protected.GET("/backups", financial.ListBackups)
	protected.POST("/backups", financial.CreateBackup)
	protected.GET("/backups/:name", financial.DownloadBackup)
	protected.DELETE("/backups/:name", financial.DeleteBackup)

	protected.GET("/dashboard/summary", financial.DashboardSummary)
	protected.GET("/dashboard/chart", financial.DashboardChart)

	return engine
}
