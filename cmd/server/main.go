package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jinmu/go-blockchain/internal/api"
	"github.com/jinmu/go-blockchain/internal/app"
	"github.com/jinmu/go-blockchain/internal/config"
)

func main() {
	// 1. 加载配置
	cfg := config.Load()
	log.Printf("Starting server on %s", cfg.Server.Address())

	// 2. 初始化应用
	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// 3. 设置路由
	router := api.NewRouter(application.Manager)
	httpServer := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: router.SetupRoutes(),
	}

	// 4. 启动充值监听器（后台运行）
	ctx := context.Background()
	go func() {
		if err := application.StartDepositListeners(ctx); err != nil {
			log.Printf("Failed to start deposit listeners: %v", err)
		}
	}()

	// 5. 启动 HTTP 服务器（goroutine）
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started successfully on http://%s", cfg.Server.Address())
	log.Printf("API documentation:")
	log.Printf("  Health: GET http://%s/health", cfg.Server.Address())
	log.Printf("  Assets: POST http://%s/api/v1/assets", cfg.Server.Address())
	log.Printf("  Accounts: POST http://%s/api/v1/accounts", cfg.Server.Address())
	log.Printf("  Balances: GET http://%s/api/v1/balances/:user_id/:asset", cfg.Server.Address())
	log.Printf("  Withdrawals: POST http://%s/api/v1/withdrawals", cfg.Server.Address())
	log.Printf("  Deposits: GET http://%s/api/v1/deposits/:tx_hash", cfg.Server.Address())

	// 6. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
