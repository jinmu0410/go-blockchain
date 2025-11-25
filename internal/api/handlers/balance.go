package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// BalanceHandler 余额相关处理器
type BalanceHandler struct {
	manager *service.Manager
}

// NewBalanceHandler 创建余额处理器
func NewBalanceHandler(manager *service.Manager) *BalanceHandler {
	return &BalanceHandler{manager: manager}
}

// GetBalance 获取余额
// GET /api/v1/balances/:user_id/:asset_symbol
func (h *BalanceHandler) GetBalance(c *gin.Context) {
	userID := c.Param("user_id")
	assetSymbol := c.Param("asset_symbol")

	balance, err := h.manager.GetBalance(c.Request.Context(), userID, assetSymbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"asset":    assetSymbol,
		"balance":  balance.Available.String(),
		"frozen":   balance.Frozen.String(),
		"pending": balance.Pending.String(),
	})
}

// ListBalances 列出用户所有余额
// GET /api/v1/balances/:user_id
func (h *BalanceHandler) ListBalances(c *gin.Context) {
	userID := c.Param("user_id")
	
	// TODO: 实现 ListBalances 方法
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "Not implemented yet",
	})
}

