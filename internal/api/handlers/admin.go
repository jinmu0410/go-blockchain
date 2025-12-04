package handlers

import (
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// AdminHandler 后台管理处理器
type AdminHandler struct {
	manager *service.Manager
}

// NewAdminHandler 创建后台管理处理器
func NewAdminHandler(manager *service.Manager) *AdminHandler {
	return &AdminHandler{manager: manager}
}

// ==================== 交易管理 ====================

// ListDeposits 列出充值记录
func (h *AdminHandler) ListDeposits(c *gin.Context) {
	// TODO: 从数据库查询充值记录
	// 这里先返回空列表，需要实现查询接口
	c.JSON(http.StatusOK, gin.H{
		"deposits": []domain.DepositRecord{},
		"total":    0,
	})
}

// GetDeposit 获取充值详情
func (h *AdminHandler) GetDeposit(c *gin.Context) {
	txHash := c.Param("tx_hash")
	deposit, err := h.manager.GetDeposit(c.Request.Context(), txHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deposit)
}

// ManualCreditDeposit 手动确认充值
func (h *AdminHandler) ManualCreditDeposit(c *gin.Context) {
	txHash := c.Param("tx_hash")
	if err := h.manager.ManualCredit(c.Request.Context(), txHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "充值已确认",
		"tx_hash": txHash,
	})
}

// ListWithdrawals 列出提现记录
func (h *AdminHandler) ListWithdrawals(c *gin.Context) {
	// TODO: 从数据库查询提现记录
	// 这里先返回空列表，需要实现查询接口
	c.JSON(http.StatusOK, gin.H{
		"withdrawals": []domain.WithdrawalRequest{},
		"total":       0,
	})
}

// GetWithdrawal 获取提现详情
func (h *AdminHandler) GetWithdrawal(c *gin.Context) {
	id := c.Param("id")
	withdrawal, err := h.manager.GetWithdrawal(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, withdrawal)
}

// ApproveWithdrawal 审批提现
func (h *AdminHandler) ApproveWithdrawal(c *gin.Context) {
	id := c.Param("id")
	withdrawal, err := h.manager.GetWithdrawal(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if withdrawal.Status != domain.WithdrawalUnderReview {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "提现状态不正确，无法审批",
			"status": withdrawal.Status,
		})
		return
	}

	// TODO: 实现审批逻辑
	// 这里需要修改 Manager 添加审批方法
	c.JSON(http.StatusOK, gin.H{
		"message": "提现已审批",
		"id":      id,
	})
}

// RejectWithdrawal 拒绝提现
func (h *AdminHandler) RejectWithdrawal(c *gin.Context) {
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	withdrawal, err := h.manager.GetWithdrawal(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if withdrawal.Status != domain.WithdrawalUnderReview {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "提现状态不正确，无法拒绝",
			"status": withdrawal.Status,
		})
		return
	}

	// TODO: 实现拒绝逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "提现已拒绝",
		"id":      id,
		"reason":  req.Reason,
	})
}

// ==================== 账号管理 ====================

// ListAccounts 列出账户
func (h *AdminHandler) ListAccounts(c *gin.Context) {
	_ = c.Query("user_id")      // 保留用于后续实现
	_ = c.Query("asset_symbol") // 保留用于后续实现
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// TODO: 实现账户列表查询
	// 需要添加 Manager.ListAccounts 方法
	c.JSON(http.StatusOK, gin.H{
		"accounts": []domain.WalletAccount{},
		"total":    0,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetAccount 获取账户详情
func (h *AdminHandler) GetAccount(c *gin.Context) {
	userID := c.Param("user_id")
	assetSymbol := c.Param("asset_symbol")

	account, err := h.manager.GetAccount(c.Request.Context(), userID, assetSymbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 获取余额
	balance, _ := h.manager.GetBalance(c.Request.Context(), userID, assetSymbol)

	c.JSON(http.StatusOK, gin.H{
		"account": account,
		"balance": balance,
	})
}

// GetBalance 获取余额
func (h *AdminHandler) GetBalance(c *gin.Context) {
	userID := c.Param("user_id")
	assetSymbol := c.Param("asset_symbol")

	balance, err := h.manager.GetBalance(c.Request.Context(), userID, assetSymbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// AdjustBalance 调整余额（管理员操作）
func (h *AdminHandler) AdjustBalance(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		AssetSymbol string `json:"asset_symbol" binding:"required"`
		Type        string `json:"type" binding:"required"` // credit, debit, freeze, unfreeze
		Amount      string `json:"amount" binding:"required"`
		Reason      string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的金额"})
		return
	}

	ctx := c.Request.Context()
	var err error

	switch req.Type {
	case "credit":
		err = h.manager.GetBalanceRepository().Credit(ctx, req.UserID, req.AssetSymbol, amount)
	case "debit":
		err = h.manager.GetBalanceRepository().Debit(ctx, req.UserID, req.AssetSymbol, amount)
	case "freeze":
		err = h.manager.GetBalanceRepository().Freeze(ctx, req.UserID, req.AssetSymbol, amount)
	case "unfreeze":
		err = h.manager.GetBalanceRepository().Unfreeze(ctx, req.UserID, req.AssetSymbol, amount)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的操作类型"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "余额调整成功",
		"type":    req.Type,
		"amount":  req.Amount,
		"reason":  req.Reason,
	})
}

// ==================== 资产管理 ====================

// ListAssets 列出资产
func (h *AdminHandler) ListAssets(c *gin.Context) {
	assets, err := h.manager.ListAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"assets": assets,
		"total":  len(assets),
	})
}

// RegisterAsset 注册资产
func (h *AdminHandler) RegisterAsset(c *gin.Context) {
	var asset domain.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.manager.RegisterAsset(c.Request.Context(), asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "资产注册成功",
		"asset":   asset,
	})
}

// ==================== 统计信息 ====================

// GetStatistics 获取统计信息
// GET /admin/api/v1/statistics?start_time=2024-01-01T00:00:00Z&end_time=2024-01-31T23:59:59Z
func (h *AdminHandler) GetStatistics(c *gin.Context) {
	// 解析时间参数
	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	// 获取充值统计
	depositStats, err := h.manager.GetDepositStatistics(c.Request.Context(), startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取资产列表
	assets, _ := h.manager.ListAssets(c.Request.Context())

	// 计算今日统计
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := time.Now()
	todayStats, _ := h.manager.GetDepositStatistics(c.Request.Context(), &todayStart, &todayEnd)

	stats := gin.H{
		"total_assets":      len(assets),
		"deposits":          depositStats,
		"today_deposits":    todayStats.TotalCount,
		"today_deposit_amount": todayStats.TotalAmount,
		"updated_at":        time.Now(),
	}

	c.JSON(http.StatusOK, stats)
}

