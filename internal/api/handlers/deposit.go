package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// DepositHandler 充值相关处理器
type DepositHandler struct {
	manager *service.Manager
}

// NewDepositHandler 创建充值处理器
func NewDepositHandler(manager *service.Manager) *DepositHandler {
	return &DepositHandler{manager: manager}
}

// GetDeposit 获取充值记录
// GET /api/v1/deposits/:tx_hash
func (h *DepositHandler) GetDeposit(c *gin.Context) {
	txHash := c.Param("tx_hash")

	deposit, err := h.manager.GetDeposit(c.Request.Context(), txHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deposit)
}

// ListDeposits 列出充值记录
// GET /api/v1/deposits?user_id=xxx&asset=xxx&status=xxx
func (h *DepositHandler) ListDeposits(c *gin.Context) {
	userID := c.Query("user_id")
	asset := c.Query("asset")
	status := c.Query("status")

	// TODO: 实现 ListDeposits 方法
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"asset":   asset,
		"status":  status,
		"message": "Not implemented yet",
	})
}

