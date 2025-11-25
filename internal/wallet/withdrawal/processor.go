package withdrawal

import (
	"context"
	"math/big"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
)

// HotWallet 表示可用于提现的热钱包
type HotWallet struct {
	Address     string
	AssetSymbol string
	Chain       domain.ChainType
	DailyLimit  *big.Int
	UsedAmount  *big.Int
	Priority    int
	Metadata    map[string]string
}

// Selector 选择合适热钱包
type Selector interface {
	Select(ctx context.Context, req domain.WithdrawalRequest) (HotWallet, error)
	UpdateUsage(ctx context.Context, wallet HotWallet, amount *big.Int) error
}

// NonceManager 管理链上交易 Nonce, 避免冲突
type NonceManager interface {
	Current(ctx context.Context, chain domain.ChainType, address string) (uint64, error)
	Increase(ctx context.Context, chain domain.ChainType, address string) error
}

// GasEstimator 获取 Gas 费率
type GasEstimator interface {
	Estimate(ctx context.Context, chain domain.ChainType, payload map[string]interface{}) (Estimation, error)
}

// Estimation 描述链上费用信息
type Estimation struct {
	BaseFeePerGas     *big.Int
	PriorityFeePerGas *big.Int
	MaxFeePerGas      *big.Int
	BlockNumber       uint64
	Metadata          map[string]interface{}
}

// BatchBuilder 适配 EIP-7702 等批量交易
type BatchBuilder interface {
	Build(ctx context.Context, requests []domain.WithdrawalRequest) ([]byte, error)
}

// Processor 负责提现执行策略
type Processor struct {
	manager  *service.Manager
	selector Selector
	nonces   NonceManager
	gas      GasEstimator
	batch    BatchBuilder
}

// NewProcessor 构造提现处理器
func NewProcessor(manager *service.Manager, selector Selector, nonces NonceManager, gas GasEstimator, batch BatchBuilder) *Processor {
	return &Processor{manager: manager, selector: selector, nonces: nonces, gas: gas, batch: batch}
}

// Process 单笔提现
func (p *Processor) Process(ctx context.Context, req domain.WithdrawalRequest) (domain.WithdrawalRequest, error) {
	wallet, err := p.selector.Select(ctx, req)
	if err != nil {
		return domain.WithdrawalRequest{}, err
	}
	if err := p.selector.UpdateUsage(ctx, wallet, req.Amount); err != nil {
		return domain.WithdrawalRequest{}, err
	}
	if p.nonces != nil {
		nonce, err := p.nonces.Current(ctx, wallet.Chain, wallet.Address)
		if err != nil {
			return domain.WithdrawalRequest{}, err
		}
		req.Metadata = mergeStringMap(req.Metadata, map[string]string{"nonce": big.NewInt(int64(nonce)).String()})
		if err := p.nonces.Increase(ctx, wallet.Chain, wallet.Address); err != nil {
			return domain.WithdrawalRequest{}, err
		}
	}
	if p.gas != nil {
		estimation, err := p.gas.Estimate(ctx, wallet.Chain, map[string]interface{}{"type": "withdrawal"})
		if err == nil {
			req.Metadata = mergeStringMap(req.Metadata, map[string]string{
				"gas_base":     estimation.BaseFeePerGas.String(),
				"gas_priority": estimation.PriorityFeePerGas.String(),
			})
		}
	}
	return p.manager.CreateWithdrawal(ctx, req)
}

// ProcessBatch 批量提现
func (p *Processor) ProcessBatch(ctx context.Context, requests []domain.WithdrawalRequest) ([]domain.WithdrawalRequest, error) {
	if p.batch != nil {
		if _, err := p.batch.Build(ctx, requests); err != nil {
			return nil, err
		}
	}
	results := make([]domain.WithdrawalRequest, 0, len(requests))
	for _, req := range requests {
		result, err := p.Process(ctx, req)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func mergeStringMap(target map[string]string, source map[string]string) map[string]string {
	if target == nil {
		target = make(map[string]string)
	}
	for k, v := range source {
		target[k] = v
	}
	return target
}
