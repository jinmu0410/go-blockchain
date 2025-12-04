package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
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
// GET /api/v1/deposits?user_id=xxx&asset=xxx&status=xxx&limit=20&offset=0
func (h *DepositHandler) ListDeposits(c *gin.Context) {
	userID := c.Query("user_id")
	asset := c.Query("asset")
	statusStr := c.Query("status")
	
	// 解析分页参数
	limit := 20 // 默认每页20条
	offset := 0  // 默认从第0条开始
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit := parseInt(limitStr); parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset := parseInt(offsetStr); parsedOffset >= 0 {
			offset = parsedOffset
		}
	}
	
	// 解析状态
	var status domain.DepositStatus
	if statusStr != "" {
		status = domain.DepositStatus(statusStr)
		// 验证状态值
		if status != domain.DepositPending && 
		   status != domain.DepositConfirmed && 
		   status != domain.DepositCredited && 
		   status != domain.DepositFailed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
	}
	
	// 查询充值记录
	deposits, err := h.manager.ListDeposits(c.Request.Context(), userID, asset, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"deposits": deposits,
		"limit":    limit,
		"offset":   offset,
		"count":    len(deposits),
	})
}

// parseInt 解析整数，失败返回0
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

