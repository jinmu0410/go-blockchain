package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/api/handlers"
	"github.com/jinmu/go-blockchain/internal/api/middleware"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// Router 路由配置
type Router struct {
	manager *service.Manager
}

// NewRouter 创建路由
func NewRouter(manager *service.Manager) *Router {
	return &Router{manager: manager}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// 中间件
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		r.healthCheck(c)
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		r.setupAssetRoutes(v1)
		r.setupAccountRoutes(v1)
		r.setupBalanceRoutes(v1)
		r.setupDepositRoutes(v1)
		r.setupWithdrawalRoutes(v1)
		r.setupRiskRoutes(v1)
	}

	// 认证路由（必须在静态文件之前）
	auth := router.Group("/admin/api")
	{
		authHandler := handlers.NewAuthHandler()
		auth.POST("/login", authHandler.Login)
	}

	// Admin API（需要认证）
	admin := router.Group("/admin/api/v1")
	admin.Use(handlers.VerifyToken())
	{
		r.setupAdminRoutes(admin)
	}

	// 静态文件服务（管理后台前端）
	// 使用单独的路径避免与 API 路由冲突
	adminGroup := router.Group("/admin")
	{
		// 首页
		adminGroup.GET("", func(c *gin.Context) {
			c.File("./web/admin/index.html")
		})
		adminGroup.GET("/", func(c *gin.Context) {
			c.File("./web/admin/index.html")
		})
		// 静态资源（CSS、JS 等）
		adminGroup.Static("/static", "./web/admin")
	}

	return router
}

// setupAssetRoutes 设置资产路由
func (r *Router) setupAssetRoutes(group *gin.RouterGroup) {
	handler := handlers.NewAssetHandler(r.manager)
	assets := group.Group("/assets")
	{
		assets.POST("", handler.RegisterAsset)
		assets.GET("", handler.ListAssets)
		assets.GET("/:symbol", handler.GetAsset)
	}
}

// setupAccountRoutes 设置账户路由
func (r *Router) setupAccountRoutes(group *gin.RouterGroup) {
	handler := handlers.NewAccountHandler(r.manager)
	accounts := group.Group("/accounts")
	{
		accounts.POST("", handler.CreateAccount)
		accounts.GET("/:user_id/:asset_symbol", handler.GetAccount)
	}
}

// setupBalanceRoutes 设置余额路由
func (r *Router) setupBalanceRoutes(group *gin.RouterGroup) {
	handler := handlers.NewBalanceHandler(r.manager)
	balances := group.Group("/balances")
	{
		balances.GET("/:user_id/:asset_symbol", handler.GetBalance)
		balances.GET("/:user_id", handler.ListBalances)
	}
}

// setupDepositRoutes 设置充值路由
func (r *Router) setupDepositRoutes(group *gin.RouterGroup) {
	handler := handlers.NewDepositHandler(r.manager)
	deposits := group.Group("/deposits")
	{
		deposits.GET("/:tx_hash", handler.GetDeposit)
		deposits.GET("", handler.ListDeposits)
	}
}

// setupWithdrawalRoutes 设置提现路由
func (r *Router) setupWithdrawalRoutes(group *gin.RouterGroup) {
	handler := handlers.NewWithdrawalHandler(r.manager)
	withdrawals := group.Group("/withdrawals")
	{
		withdrawals.POST("", handler.CreateWithdrawal)
		withdrawals.GET("/:id", handler.GetWithdrawal)
		withdrawals.GET("", handler.ListWithdrawals)
	}
}

// setupRiskRoutes 设置风控路由
func (r *Router) setupRiskRoutes(group *gin.RouterGroup) {
	handler := handlers.NewRiskHandler(r.manager)
	risk := group.Group("/risk")
	{
		// 白名单管理
		risk.POST("/whitelist", handler.AddToWhitelist)
		risk.DELETE("/whitelist/:address", handler.RemoveFromWhitelist)
		risk.GET("/whitelist", handler.ListWhitelist)
		
		// 黑名单管理
		risk.POST("/blacklist", handler.AddToBlacklist)
		risk.DELETE("/blacklist/:address", handler.RemoveFromBlacklist)
		risk.GET("/blacklist", handler.ListBlacklist)
		
		// 配置管理
		risk.GET("/config", handler.GetRiskConfig)
		risk.PUT("/config", handler.UpdateRiskConfig)
		
		// 日志查询
		risk.GET("/logs", handler.ListRiskLogs)
	}
}

// setupAdminRoutes 设置后台管理路由
func (r *Router) setupAdminRoutes(group *gin.RouterGroup) {
	handler := handlers.NewAdminHandler(r.manager)
	
	// 交易管理
	transactions := group.Group("/transactions")
	{
		transactions.GET("/deposits", handler.ListDeposits)
		transactions.GET("/deposits/:tx_hash", handler.GetDeposit)
		transactions.POST("/deposits/:tx_hash/credit", handler.ManualCreditDeposit)
		
		transactions.GET("/withdrawals", handler.ListWithdrawals)
		transactions.GET("/withdrawals/:id", handler.GetWithdrawal)
		transactions.POST("/withdrawals/:id/approve", handler.ApproveWithdrawal)
		transactions.POST("/withdrawals/:id/reject", handler.RejectWithdrawal)
	}
	
	// 账号管理
	accounts := group.Group("/accounts")
	{
		accounts.GET("", handler.ListAccounts)
		accounts.GET("/:user_id/:asset_symbol", handler.GetAccount)
		accounts.GET("/:user_id/:asset_symbol/balance", handler.GetBalance)
		accounts.POST("/balance/adjust", handler.AdjustBalance)
	}
	
	// 资产管理
	assets := group.Group("/assets")
	{
		assets.GET("", handler.ListAssets)
		assets.POST("", handler.RegisterAsset)
	}
	
	// 统计信息
	group.GET("/statistics", handler.GetStatistics)
}

// healthCheck 健康检查
func (r *Router) healthCheck(c *gin.Context) {
	response := gin.H{
		"status":  "ok",
		"service": "wallet-api",
	}
	
	// 获取扫描器状态
	scannerStatuses := r.manager.GetScannerStatuses(c.Request.Context())
	if len(scannerStatuses) > 0 {
		response["scanners"] = scannerStatuses
		
		// 检查是否有不健康的扫描器
		allHealthy := true
		for _, status := range scannerStatuses {
			if !status.IsHealthy {
				allHealthy = false
				break
			}
		}
		if !allHealthy {
			response["status"] = "degraded"
		}
	}
	
	c.JSON(200, response)
}

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

