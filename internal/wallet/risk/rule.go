package risk

import (
	"context"
	"math/big"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// Rule 风控规则接口
type Rule interface {
	// Evaluate 评估规则，返回风险评分增量 (0-1)
	Evaluate(ctx context.Context, req domain.WithdrawalRequest, data *EvaluationData) (float64, string, error)
	// Name 返回规则名称
	Name() string
	// Priority 返回规则优先级（数字越小优先级越高）
	Priority() int
}

// EvaluationData 评估数据上下文
type EvaluationData struct {
	UserID              string
	AssetSymbol         string
	RecentWithdrawals   []domain.WithdrawalRequest // 最近的提现记录
	DailyAmount         *big.Int                    // 今日累计提现金额
	MonthlyAmount       *big.Int                    // 本月累计提现金额
	WithdrawalCount     int                         // 今日提现次数
	AccountCreatedAt    time.Time                   // 账户创建时间
	LastWithdrawalAt    *time.Time                  // 上次提现时间
	IsWhitelisted       bool                        // 是否在白名单
	IsBlacklisted       bool                        // 是否在黑名单
	AddressRiskScore    float64                     // 地址风险评分
}

// RuleEngine 规则引擎
type RuleEngine struct {
	rules []Rule
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(rules []Rule) *RuleEngine {
	return &RuleEngine{rules: rules}
}

// Evaluate 执行所有规则评估
func (e *RuleEngine) Evaluate(ctx context.Context, req domain.WithdrawalRequest, data *EvaluationData) (float64, []string, error) {
	var totalScore float64
	var remarks []string

	// 按优先级排序规则
	sortedRules := make([]Rule, len(e.rules))
	copy(sortedRules, e.rules)
	for i := 0; i < len(sortedRules)-1; i++ {
		for j := i + 1; j < len(sortedRules); j++ {
			if sortedRules[i].Priority() > sortedRules[j].Priority() {
				sortedRules[i], sortedRules[j] = sortedRules[j], sortedRules[i]
			}
		}
	}

	// 执行规则评估
	for _, rule := range sortedRules {
		score, remark, err := rule.Evaluate(ctx, req, data)
		if err != nil {
			return 0, nil, err
		}
		if score > 0 {
			totalScore += score
			if remark != "" {
				remarks = append(remarks, remark)
			}
		}
	}

	// 确保评分在 0-1 范围内
	if totalScore > 1.0 {
		totalScore = 1.0
	}

	return totalScore, remarks, nil
}

// AmountLimitRule 金额限制规则
type AmountLimitRule struct {
	SingleMaxAmount *big.Int // 单笔最大金额
	DailyMaxAmount  *big.Int // 日累计最大金额
	MonthlyMaxAmount *big.Int // 月累计最大金额
}

func NewAmountLimitRule(singleMax, dailyMax, monthlyMax *big.Int) *AmountLimitRule {
	return &AmountLimitRule{
		SingleMaxAmount:  singleMax,
		DailyMaxAmount:   dailyMax,
		MonthlyMaxAmount: monthlyMax,
	}
}

func (r *AmountLimitRule) Name() string {
	return "金额限制规则"
}

func (r *AmountLimitRule) Priority() int {
	return 1
}

func (r *AmountLimitRule) Evaluate(ctx context.Context, req domain.WithdrawalRequest, data *EvaluationData) (float64, string, error) {
	var score float64
	var remarks []string

	// 单笔金额检查
	if r.SingleMaxAmount != nil && req.Amount.Cmp(r.SingleMaxAmount) > 0 {
		score += 0.5
		remarks = append(remarks, "单笔金额超过限制")
	}

	// 日累计金额检查
	if r.DailyMaxAmount != nil && data.DailyAmount != nil {
		totalDaily := new(big.Int).Add(data.DailyAmount, req.Amount)
		if totalDaily.Cmp(r.DailyMaxAmount) > 0 {
			score += 0.4
			remarks = append(remarks, "日累计金额超过限制")
		}
	}

	// 月累计金额检查
	if r.MonthlyMaxAmount != nil && data.MonthlyAmount != nil {
		totalMonthly := new(big.Int).Add(data.MonthlyAmount, req.Amount)
		if totalMonthly.Cmp(r.MonthlyMaxAmount) > 0 {
			score += 0.3
			remarks = append(remarks, "月累计金额超过限制")
		}
	}

	remark := ""
	if len(remarks) > 0 {
		remark = remarks[0]
		if len(remarks) > 1 {
			remark += "等"
		}
	}

	return score, remark, nil
}

// FrequencyLimitRule 频率限制规则
type FrequencyLimitRule struct {
	MaxCountPerDay    int           // 每日最大提现次数
	MinInterval       time.Duration // 最小提现间隔
}

func NewFrequencyLimitRule(maxCountPerDay int, minInterval time.Duration) *FrequencyLimitRule {
	return &FrequencyLimitRule{
		MaxCountPerDay: maxCountPerDay,
		MinInterval:    minInterval,
	}
}

func (r *FrequencyLimitRule) Name() string {
	return "频率限制规则"
}

func (r *FrequencyLimitRule) Priority() int {
	return 2
}

func (r *FrequencyLimitRule) Evaluate(ctx context.Context, req domain.WithdrawalRequest, data *EvaluationData) (float64, string, error) {
	var score float64
	var remark string

	// 每日提现次数检查
	if data.WithdrawalCount >= r.MaxCountPerDay {
		score += 0.4
		remark = "每日提现次数超过限制"
	}

	// 提现间隔检查
	if data.LastWithdrawalAt != nil {
		interval := time.Since(*data.LastWithdrawalAt)
		if interval < r.MinInterval {
			score += 0.3
			if remark != "" {
				remark += "; "
			}
			remark += "提现间隔过短"
		}
	}

	return score, remark, nil
}

// AddressRule 地址规则（白名单/黑名单）
type AddressRule struct {
	whitelist map[string]bool
	blacklist map[string]bool
}

func NewAddressRule() *AddressRule {
	return &AddressRule{
		whitelist: make(map[string]bool),
		blacklist: make(map[string]bool),
	}
}

func (r *AddressRule) Name() string {
	return "地址规则"
}

func (r *AddressRule) Priority() int {
	return 0 // 最高优先级
}

func (r *AddressRule) Evaluate(ctx context.Context, req domain.WithdrawalRequest, data *EvaluationData) (float64, string, error) {
	// 黑名单检查 - 直接拒绝
	if data.IsBlacklisted || r.blacklist[req.ToAddress] {
		return 1.0, "目标地址在黑名单中", nil
	}

	// 白名单检查 - 降低风险评分
	if data.IsWhitelisted || r.whitelist[req.ToAddress] {
		return -0.2, "目标地址在白名单中", nil // 负数表示降低风险
	}

	return 0, "", nil
}

// AddToWhitelist 添加到白名单
func (r *AddressRule) AddToWhitelist(address string) {
	r.whitelist[address] = true
}

// RemoveFromWhitelist 从白名单移除
func (r *AddressRule) RemoveFromWhitelist(address string) {
	delete(r.whitelist, address)
}

// AddToBlacklist 添加到黑名单
func (r *AddressRule) AddToBlacklist(address string) {
	r.blacklist[address] = true
}

// RemoveFromBlacklist 从黑名单移除
func (r *AddressRule) RemoveFromBlacklist(address string) {
	delete(r.blacklist, address)
}

// AccountAgeRule 账户年龄规则
type AccountAgeRule struct {
	MinAge time.Duration // 最小账户年龄
}

func NewAccountAgeRule(minAge time.Duration) *AccountAgeRule {
	return &AccountAgeRule{MinAge: minAge}
}

func (r *AccountAgeRule) Name() string {
	return "账户年龄规则"
}

func (r *AccountAgeRule) Priority() int {
	return 3
}

func (r *AccountAgeRule) Evaluate(ctx context.Context, req domain.WithdrawalRequest, data *EvaluationData) (float64, string, error) {
	if data.AccountCreatedAt.IsZero() {
		return 0, "", nil
	}

	age := time.Since(data.AccountCreatedAt)
	if age < r.MinAge {
		// 账户越新，风险越高
		riskRatio := 1.0 - (age.Seconds() / r.MinAge.Seconds())
		if riskRatio < 0 {
			riskRatio = 0
		}
		return riskRatio * 0.3, "账户创建时间过短", nil
	}

	return 0, "", nil
}

// AddressRiskRule 地址风险评分规则
type AddressRiskRule struct {
	HighRiskThreshold float64 // 高风险阈值
}

func NewAddressRiskRule(threshold float64) *AddressRiskRule {
	return &AddressRiskRule{HighRiskThreshold: threshold}
}

func (r *AddressRiskRule) Name() string {
	return "地址风险评分规则"
}

func (r *AddressRiskRule) Priority() int {
	return 4
}

func (r *AddressRiskRule) Evaluate(ctx context.Context, req domain.WithdrawalRequest, data *EvaluationData) (float64, string, error) {
	if data.AddressRiskScore >= r.HighRiskThreshold {
		return data.AddressRiskScore * 0.5, "目标地址风险评分较高", nil
	}

	return 0, "", nil
}

