package risk

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/store"
)

// RiskController 风控控制器实现
type RiskController struct {
	engine      *RuleEngine
	store       store.RepositoryProvider
	config      *Config
	addressRule *AddressRule
	addressList AddressListRepository
	riskLog     RiskLogRepository
}

// Config 风控配置
type Config struct {
	// 金额限制
	SingleMaxAmount  *big.Int // 单笔最大金额（nil 表示不限制）
	DailyMaxAmount   *big.Int // 日累计最大金额
	MonthlyMaxAmount *big.Int // 月累计最大金额

	// 频率限制
	MaxCountPerDay int           // 每日最大提现次数
	MinInterval    time.Duration // 最小提现间隔

	// 账户年龄
	MinAccountAge time.Duration // 最小账户年龄

	// 风险评分阈值
	AutoApproveScore  float64 // 自动通过阈值（低于此分数自动通过）
	ManualReviewScore float64 // 人工审核阈值（高于此分数需要人工审核）
	RejectScore       float64 // 拒绝阈值（高于此分数直接拒绝）

	// 地址风险
	AddressRiskThreshold float64 // 地址风险阈值
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	ethUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil) // 1 ETH

	return &Config{
		SingleMaxAmount:      new(big.Int).Mul(big.NewInt(100), ethUnit),   // 100 ETH
		DailyMaxAmount:       new(big.Int).Mul(big.NewInt(1000), ethUnit),  // 1000 ETH
		MonthlyMaxAmount:     new(big.Int).Mul(big.NewInt(10000), ethUnit), // 10000 ETH
		MaxCountPerDay:       10,
		MinInterval:          5 * time.Minute,
		MinAccountAge:        24 * time.Hour,
		AutoApproveScore:     0.3,
		ManualReviewScore:    0.5,
		RejectScore:          0.7,
		AddressRiskThreshold: 0.6,
	}
}

// NewRiskController 创建风控控制器
func NewRiskController(store store.RepositoryProvider, config *Config) *RiskController {
	return NewRiskControllerWithRepos(store, config, NewInMemoryAddressListStore(), NewInMemoryRiskLogStore())
}

// NewRiskControllerWithRepos 创建风控控制器（带自定义存储）
func NewRiskControllerWithRepos(store store.RepositoryProvider, config *Config, addressList AddressListRepository, riskLog RiskLogRepository) *RiskController {
	if config == nil {
		config = DefaultConfig()
	}

	addressRule := NewAddressRule()

	// 初始化规则引擎
	rules := []Rule{
		addressRule, // 地址规则（最高优先级）
		NewAmountLimitRule(config.SingleMaxAmount, config.DailyMaxAmount, config.MonthlyMaxAmount),
		NewFrequencyLimitRule(config.MaxCountPerDay, config.MinInterval),
		NewAccountAgeRule(config.MinAccountAge),
		NewAddressRiskRule(config.AddressRiskThreshold),
	}

	engine := NewRuleEngine(rules)

	return &RiskController{
		engine:      engine,
		store:       store,
		config:      config,
		addressRule: addressRule,
		addressList: addressList,
		riskLog:     riskLog,
	}
}

// EvaluateWithdrawal 评估提现请求
func (c *RiskController) EvaluateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (domain.WithdrawalDecision, error) {
	// 1. 收集评估数据
	data, err := c.collectEvaluationData(ctx, req)
	if err != nil {
		return domain.WithdrawalDecision{
			Approved: false,
			Score:    1.0,
			Remarks:  fmt.Sprintf("数据收集失败: %v", err),
		}, err
	}

	// 2. 执行规则评估
	score, remarks, err := c.engine.Evaluate(ctx, req, data)
	if err != nil {
		return domain.WithdrawalDecision{
			Approved: false,
			Score:    1.0,
			Remarks:  fmt.Sprintf("规则评估失败: %v", err),
		}, err
	}

	// 3. 处理负数评分（白名单等降低风险的情况）
	if score < 0 {
		score = 0
	}

	// 4. 根据评分做出决策
	decision := c.makeDecision(score, remarks)

	// 5. 记录风控日志（异步）
	go c.logDecision(ctx, req, decision, data)

	return decision, nil
}

// collectEvaluationData 收集评估数据
func (c *RiskController) collectEvaluationData(ctx context.Context, req domain.WithdrawalRequest) (*EvaluationData, error) {
	data := &EvaluationData{
		UserID:      req.UserID,
		AssetSymbol: req.AssetSymbol,
	}

	// 获取账户信息
	account, err := c.store.GetAccount(ctx, req.UserID, req.AssetSymbol)
	if err == nil {
		data.AccountCreatedAt = account.CreatedAt
	}

	// 获取最近的提现记录
	// TODO: 实现查询最近提现记录的方法
	// 这里先使用空列表，实际应该从数据库查询
	data.RecentWithdrawals = []domain.WithdrawalRequest{}

	// 计算今日和本月的累计金额和次数
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var dailyAmount, monthlyAmount *big.Int
	var withdrawalCount int
	var lastWithdrawalAt *time.Time

	// 遍历最近的提现记录（这里简化处理，实际应该从数据库查询）
	for _, w := range data.RecentWithdrawals {
		if w.CreatedAt.After(todayStart) {
			if dailyAmount == nil {
				dailyAmount = big.NewInt(0)
			}
			dailyAmount = new(big.Int).Add(dailyAmount, w.Amount)
			withdrawalCount++
		}
		if w.CreatedAt.After(monthStart) {
			if monthlyAmount == nil {
				monthlyAmount = big.NewInt(0)
			}
			monthlyAmount = new(big.Int).Add(monthlyAmount, w.Amount)
		}
		if lastWithdrawalAt == nil || w.CreatedAt.After(*lastWithdrawalAt) {
			t := w.CreatedAt
			lastWithdrawalAt = &t
		}
	}

	data.DailyAmount = dailyAmount
	data.MonthlyAmount = monthlyAmount
	data.WithdrawalCount = withdrawalCount
	data.LastWithdrawalAt = lastWithdrawalAt

	// 检查白名单/黑名单（优先从存储查询）
	if c.addressList != nil {
		data.IsWhitelisted, _ = c.addressList.IsWhitelisted(ctx, req.ToAddress, string(req.Chain))
		data.IsBlacklisted, _ = c.addressList.IsBlacklisted(ctx, req.ToAddress, string(req.Chain))
	} else {
		// 回退到内存规则
		data.IsWhitelisted = c.addressRule.whitelist[req.ToAddress]
		data.IsBlacklisted = c.addressRule.blacklist[req.ToAddress]
	}

	// 地址风险评分（简化实现，实际可以调用外部服务）
	data.AddressRiskScore = c.evaluateAddressRisk(req.ToAddress)

	return data, nil
}

// evaluateAddressRisk 评估地址风险（简化实现）
func (c *RiskController) evaluateAddressRisk(address string) float64 {
	// 这里可以实现更复杂的地址风险评分逻辑
	// 例如：检查地址是否在已知的恶意地址列表中
	// 或者调用外部风险评分服务

	// 简化实现：返回一个基础评分
	// 实际应该根据地址的历史交易、关联账户等信息计算
	return 0.0
}

// makeDecision 根据评分做出决策
func (c *RiskController) makeDecision(score float64, remarks []string) domain.WithdrawalDecision {
	decision := domain.WithdrawalDecision{
		Score:    score,
		Remarks:  strings.Join(remarks, "; "),
		Metadata: make(map[string]string),
	}

	if score >= c.config.RejectScore {
		decision.Approved = false
		decision.Remarks = "风险评分过高，拒绝提现: " + decision.Remarks
		decision.Metadata["action"] = "rejected"
	} else if score >= c.config.ManualReviewScore {
		decision.Approved = false // 需要人工审核
		decision.Remarks = "需要人工审核: " + decision.Remarks
		decision.Metadata["action"] = "manual_review"
	} else if score >= c.config.AutoApproveScore {
		decision.Approved = false // 需要人工审核
		decision.Remarks = "建议人工审核: " + decision.Remarks
		decision.Metadata["action"] = "suggest_review"
	} else {
		decision.Approved = true
		decision.Remarks = "自动通过"
		decision.Metadata["action"] = "auto_approved"
	}

	decision.Metadata["score"] = fmt.Sprintf("%.2f", score)
	decision.Metadata["threshold_auto"] = fmt.Sprintf("%.2f", c.config.AutoApproveScore)
	decision.Metadata["threshold_review"] = fmt.Sprintf("%.2f", c.config.ManualReviewScore)
	decision.Metadata["threshold_reject"] = fmt.Sprintf("%.2f", c.config.RejectScore)

	return decision
}

// logDecision 记录风控决策日志
func (c *RiskController) logDecision(ctx context.Context, req domain.WithdrawalRequest, decision domain.WithdrawalDecision, data *EvaluationData) {
	if c.riskLog == nil {
		return
	}

	log := &RiskLog{
		WithdrawalID: req.ID,
		UserID:       req.UserID,
		AssetSymbol:  req.AssetSymbol,
		ToAddress:    req.ToAddress,
		Amount:       req.Amount.String(),
		Score:        decision.Score,
		Approved:     decision.Approved,
		Remarks:      decision.Remarks,
		Decision:     decision.Metadata["action"],
		Metadata:     decision.Metadata,
		Rules:        make(map[string]interface{}),
	}

	// 记录规则详情
	log.Rules["daily_amount"] = data.DailyAmount
	log.Rules["monthly_amount"] = data.MonthlyAmount
	log.Rules["withdrawal_count"] = data.WithdrawalCount
	log.Rules["is_whitelisted"] = data.IsWhitelisted
	log.Rules["is_blacklisted"] = data.IsBlacklisted
	log.Rules["address_risk_score"] = data.AddressRiskScore

	_ = c.riskLog.SaveLog(ctx, log)
}

// AddToWhitelist 添加地址到白名单
func (c *RiskController) AddToWhitelist(ctx context.Context, address string, chain string, remark string) error {
	if c.addressList != nil {
		if err := c.addressList.AddToWhitelist(ctx, address, chain, remark); err != nil {
			return err
		}
	}
	c.addressRule.AddToWhitelist(address)
	return nil
}

// RemoveFromWhitelist 从白名单移除地址
func (c *RiskController) RemoveFromWhitelist(ctx context.Context, address string, chain string) error {
	if c.addressList != nil {
		if err := c.addressList.RemoveFromWhitelist(ctx, address, chain); err != nil {
			return err
		}
	}
	c.addressRule.RemoveFromWhitelist(address)
	return nil
}

// AddToBlacklist 添加地址到黑名单
func (c *RiskController) AddToBlacklist(ctx context.Context, address string, chain string, remark string) error {
	if c.addressList != nil {
		if err := c.addressList.AddToBlacklist(ctx, address, chain, remark); err != nil {
			return err
		}
	}
	c.addressRule.AddToBlacklist(address)
	return nil
}

// RemoveFromBlacklist 从黑名单移除地址
func (c *RiskController) RemoveFromBlacklist(ctx context.Context, address string, chain string) error {
	if c.addressList != nil {
		if err := c.addressList.RemoveFromBlacklist(ctx, address, chain); err != nil {
			return err
		}
	}
	c.addressRule.RemoveFromBlacklist(address)
	return nil
}

// GetAddressListRepository 获取地址列表存储接口
func (c *RiskController) GetAddressListRepository() AddressListRepository {
	return c.addressList
}

// GetRiskLogRepository 获取风控日志存储接口
func (c *RiskController) GetRiskLogRepository() RiskLogRepository {
	return c.riskLog
}

// GetConfig 获取当前配置
func (c *RiskController) GetConfig() *Config {
	return c.config
}

// UpdateConfig 更新配置
func (c *RiskController) UpdateConfig(config *Config) {
	c.config = config
	// 重新创建规则引擎
	rules := []Rule{
		c.addressRule,
		NewAmountLimitRule(config.SingleMaxAmount, config.DailyMaxAmount, config.MonthlyMaxAmount),
		NewFrequencyLimitRule(config.MaxCountPerDay, config.MinInterval),
		NewAccountAgeRule(config.MinAccountAge),
		NewAddressRiskRule(config.AddressRiskThreshold),
	}
	c.engine = NewRuleEngine(rules)
}
