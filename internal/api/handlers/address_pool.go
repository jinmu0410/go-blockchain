package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// AddressPoolHandler 地址池管理处理器
type AddressPoolHandler struct {
	manager *service.Manager
}

// NewAddressPoolHandler 创建地址池处理器
func NewAddressPoolHandler(manager *service.Manager) *AddressPoolHandler {
	return &AddressPoolHandler{manager: manager}
}

// ListAddresses 列出地址池中的地址
// GET /admin/api/v1/address-pool
func (h *AddressPoolHandler) ListAddresses(c *gin.Context) {
	chain := domain.ChainType(c.Query("chain"))
	assetSymbol := c.Query("asset")
	statusStr := c.Query("status")
	
	var status domain.AddressPoolStatus
	if statusStr != "" {
		status = domain.AddressPoolStatus(statusStr)
	}
	
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	addresses, err := h.manager.ListAddressPool(c.Request.Context(), chain, assetSymbol, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"addresses": addresses,
		"count":     len(addresses),
	})
}

// GetStats 获取地址池统计信息
// GET /admin/api/v1/address-pool/stats
func (h *AddressPoolHandler) GetStats(c *gin.Context) {
	chain := domain.ChainType(c.Query("chain"))
	assetSymbol := c.Query("asset")
	
	stats, err := h.manager.GetAddressPoolStats(c.Request.Context(), chain, assetSymbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}

// GenerateAddresses 批量生成地址
// POST /admin/api/v1/address-pool/generate
func (h *AddressPoolHandler) GenerateAddresses(c *gin.Context) {
	var req struct {
		Chain       string `json:"chain" binding:"required"`
		AssetSymbol string `json:"asset_symbol" binding:"required"`
		Count       int    `json:"count" binding:"required,min=1,max=1000"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 验证链类型
	chain := domain.ChainType(req.Chain)
	if chain != domain.ChainEVM && chain != domain.ChainBitcoin && chain != domain.ChainSolana {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chain type, must be evm, bitcoin, or solana"})
		return
	}
	
	// 先检查资产是否存在
	_, err := h.manager.GetAsset(c.Request.Context(), req.AssetSymbol)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("asset '%s' not found, please register it first", req.AssetSymbol),
		})
		return
	}
	
	if err := h.manager.GenerateAddressBatch(c.Request.Context(), chain, req.AssetSymbol, req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Addresses generated successfully",
		"count":   req.Count,
		"chain":   req.Chain,
		"asset":   req.AssetSymbol,
	})
}

