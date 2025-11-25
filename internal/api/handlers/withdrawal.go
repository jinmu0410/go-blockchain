package handlers

import (
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// WithdrawalHandler 提现相关处理器
type WithdrawalHandler struct {
	manager *service.Manager
}

// NewWithdrawalHandler 创建提现处理器
func NewWithdrawalHandler(manager *service.Manager) *WithdrawalHandler {
	return &WithdrawalHandler{manager: manager}
}

// CreateWithdrawal 创建提现请求
// POST /api/v1/withdrawals
func (h *WithdrawalHandler) CreateWithdrawal(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		AssetSymbol string `json:"asset_symbol" binding:"required"`
		ToAddress   string `json:"to_address" binding:"required"`
		Amount      string `json:"amount" binding:"required"`
		Chain       string `json:"chain,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		return
	}

	chain := domain.ChainEVM
	if req.Chain != "" {
		chain = domain.ChainType(req.Chain)
	}

	withdrawalReq := domain.WithdrawalRequest{
		UserID:      req.UserID,
		AssetSymbol: req.AssetSymbol,
		ToAddress:   req.ToAddress,
		Amount:       amount,
		Chain:       chain,
		Status:      domain.WithdrawalRequested,
		Metadata:    make(map[string]string),
	}

	result, err := h.manager.CreateWithdrawal(c.Request.Context(), withdrawalReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetWithdrawal 获取提现信息
// GET /api/v1/withdrawals/:id
func (h *WithdrawalHandler) GetWithdrawal(c *gin.Context) {
	id := c.Param("id")

	withdrawal, err := h.manager.GetWithdrawal(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, withdrawal)
}

// ListWithdrawals 列出提现记录
// GET /api/v1/withdrawals?user_id=xxx&status=xxx
func (h *WithdrawalHandler) ListWithdrawals(c *gin.Context) {
	userID := c.Query("user_id")
	statusStr := c.Query("status")

	// TODO: 实现 ListWithdrawals 方法
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"status":  statusStr,
		"message": "Not implemented yet",
	})
}

