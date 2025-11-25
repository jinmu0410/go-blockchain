package risk

import (
	"context"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// Controller 风控接口
type Controller interface {
	EvaluateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (domain.WithdrawalDecision, error)
}

// NoopController 默认通过所有提现
type NoopController struct{}

// EvaluateWithdrawal 实现风控接口
func (NoopController) EvaluateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (domain.WithdrawalDecision, error) {
	return domain.WithdrawalDecision{Approved: true, Score: 1.0, Remarks: "auto-approved"}, nil
}
