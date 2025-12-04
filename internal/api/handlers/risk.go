package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/risk"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// RiskHandler 风控管理处理器
type RiskHandler struct {
	manager *service.Manager
}

// NewRiskHandler 创建风控处理器
func NewRiskHandler(manager *service.Manager) *RiskHandler {
	return &RiskHandler{manager: manager}
}

// AddToWhitelist 添加地址到白名单
func (h *RiskHandler) AddToWhitelist(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
		Chain   string `json:"chain" binding:"required"`
		Remark  string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	if err := riskCtrl.AddToWhitelist(c.Request.Context(), req.Address, req.Chain, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "已添加到白名单",
		"address": req.Address,
		"chain":   req.Chain,
	})
}

// RemoveFromWhitelist 从白名单移除地址
func (h *RiskHandler) RemoveFromWhitelist(c *gin.Context) {
	address := c.Param("address")
	chain := c.Query("chain")
	if chain == "" {
		chain = "evm" // 默认值
	}

	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	if err := riskCtrl.RemoveFromWhitelist(c.Request.Context(), address, chain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "已从白名单移除",
		"address": address,
		"chain":   chain,
	})
}

// AddToBlacklist 添加地址到黑名单
func (h *RiskHandler) AddToBlacklist(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
		Chain   string `json:"chain" binding:"required"`
		Remark  string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	if err := riskCtrl.AddToBlacklist(c.Request.Context(), req.Address, req.Chain, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "已添加到黑名单",
		"address": req.Address,
		"chain":   req.Chain,
	})
}

// RemoveFromBlacklist 从黑名单移除地址
func (h *RiskHandler) RemoveFromBlacklist(c *gin.Context) {
	address := c.Param("address")
	chain := c.Query("chain")
	if chain == "" {
		chain = "evm" // 默认值
	}

	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	if err := riskCtrl.RemoveFromBlacklist(c.Request.Context(), address, chain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "已从黑名单移除",
		"address": address,
		"chain":   chain,
	})
}

// ListWhitelist 列出白名单
func (h *RiskHandler) ListWhitelist(c *gin.Context) {
	chain := c.Query("chain")
	if chain == "" {
		chain = "evm"
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	riskController, ok := riskCtrl.(*risk.RiskController)
	if !ok {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "当前风控实现不支持列表查询"})
		return
	}

	repo := riskController.GetAddressListRepository()
	if repo == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "地址列表存储未配置"})
		return
	}

	entries, err := repo.ListWhitelist(c.Request.Context(), chain, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chain":   chain,
		"count":   len(entries),
		"entries": entries,
	})
}

// ListBlacklist 列出黑名单
func (h *RiskHandler) ListBlacklist(c *gin.Context) {
	chain := c.Query("chain")
	if chain == "" {
		chain = "evm"
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	riskController, ok := riskCtrl.(*risk.RiskController)
	if !ok {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "当前风控实现不支持列表查询"})
		return
	}

	repo := riskController.GetAddressListRepository()
	if repo == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "地址列表存储未配置"})
		return
	}

	entries, err := repo.ListBlacklist(c.Request.Context(), chain, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chain":   chain,
		"count":   len(entries),
		"entries": entries,
	})
}

// GetRiskConfig 获取风控配置
func (h *RiskHandler) GetRiskConfig(c *gin.Context) {
	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	config := riskCtrl.GetConfig()
	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "配置不存在"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateRiskConfig 更新风控配置
func (h *RiskHandler) UpdateRiskConfig(c *gin.Context) {
	var config risk.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	riskCtrl.UpdateConfig(&config)

	c.JSON(http.StatusOK, gin.H{
		"message": "配置已更新",
		"config":  config,
	})
}

// ListRiskLogs 列出风控日志
func (h *RiskHandler) ListRiskLogs(c *gin.Context) {
	riskCtrl := h.getRiskController()
	if riskCtrl == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控功能未启用"})
		return
	}

	riskController, ok := riskCtrl.(*risk.RiskController)
	if !ok {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "当前风控实现不支持日志查询"})
		return
	}

	repo := riskController.GetRiskLogRepository()
	if repo == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "风控日志存储未配置"})
		return
	}

	filter := &risk.RiskLogFilter{
		UserID:      c.Query("user_id"),
		WithdrawalID: c.Query("withdrawal_id"),
		Limit:       100,
		Offset:      0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	logs, err := repo.ListLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"logs":  logs,
	})
}

// getRiskController 获取风控控制器
func (h *RiskHandler) getRiskController() risk.Controller {
	return h.manager.GetRiskController()
}

