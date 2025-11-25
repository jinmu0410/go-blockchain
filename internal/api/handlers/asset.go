package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// AssetHandler 资产相关处理器
type AssetHandler struct {
	manager *service.Manager
}

// NewAssetHandler 创建资产处理器
func NewAssetHandler(manager *service.Manager) *AssetHandler {
	return &AssetHandler{manager: manager}
}

// RegisterAsset 注册资产
// POST /api/v1/assets
func (h *AssetHandler) RegisterAsset(c *gin.Context) {
	var req struct {
		Symbol   string            `json:"symbol" binding:"required"`
		Chain    domain.ChainType  `json:"chain" binding:"required"`
		Decimals uint8             `json:"decimals" binding:"required"`
		TokenAddr string           `json:"token_addr,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset := domain.Asset{
		Symbol:    req.Symbol,
		Chain:     req.Chain,
		Decimals:  req.Decimals,
		TokenAddr: req.TokenAddr,
	}

	if err := h.manager.RegisterAsset(c.Request.Context(), asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asset registered successfully",
		"asset":   asset,
	})
}

// GetAsset 获取资产信息
// GET /api/v1/assets/:symbol
func (h *AssetHandler) GetAsset(c *gin.Context) {
	symbol := c.Param("symbol")

	asset, err := h.manager.GetAsset(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// ListAssets 列出所有资产
// GET /api/v1/assets
func (h *AssetHandler) ListAssets(c *gin.Context) {
	// TODO: 实现 ListAssets 方法
	c.JSON(http.StatusOK, gin.H{"message": "Not implemented yet"})
}

