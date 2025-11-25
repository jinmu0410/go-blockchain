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
	router.GET("/health", healthCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		r.setupAssetRoutes(v1)
		r.setupAccountRoutes(v1)
		r.setupBalanceRoutes(v1)
		r.setupDepositRoutes(v1)
		r.setupWithdrawalRoutes(v1)
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

// healthCheck 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "wallet-api",
	})
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

