package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// AccountHandler 账户相关处理器
type AccountHandler struct {
	manager *service.Manager
}

// NewAccountHandler 创建账户处理器
func NewAccountHandler(manager *service.Manager) *AccountHandler {
	return &AccountHandler{manager: manager}
}

// CreateAccount 创建账户
// POST /api/v1/accounts
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		AssetSymbol string `json:"asset_symbol" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.manager.EnsureAccount(c.Request.Context(), req.UserID, req.AssetSymbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// GetAccount 获取账户信息
// GET /api/v1/accounts/:user_id/:asset_symbol
func (h *AccountHandler) GetAccount(c *gin.Context) {
	userID := c.Param("user_id")
	assetSymbol := c.Param("asset_symbol")

	account, err := h.manager.GetAccount(c.Request.Context(), userID, assetSymbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

