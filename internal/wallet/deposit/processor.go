package deposit

import (
	"context"
	"math/big"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// BloomFilter 用于快速判断充值交易是否为关注地址
type BloomFilter interface {
	Add(item []byte)
	Test(item []byte) bool
}

// Ledger 资金流水表接口, 记录每一笔充值/提现以便追溯和回滚
type Ledger interface {
	AppendDeposit(ctx context.Context, record domain.DepositRecord) error
	AppendWithdrawal(ctx context.Context, req domain.WithdrawalRequest) error
	Rollback(ctx context.Context, reference string) error
}

// ReorgHandler 处理区块重组
type ReorgHandler interface {
	OnReorg(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) error
}

// Consumer 消费充值事件
type Consumer interface {
	HandleDepositEvent(ctx context.Context, event domain.DepositEvent) error
}

// Processor 聚合充值相关能力
type Processor struct {
	consumer      Consumer
	bloom         BloomFilter
	ledger        Ledger
	reorgHandler  ReorgHandler
	confirmations uint64
}

// NewProcessor 创建新的充值流程处理器
func NewProcessor(consumer Consumer, bloom BloomFilter, ledger Ledger, reorgHandler ReorgHandler, confirmations uint64) *Processor {
	return &Processor{
		consumer:      consumer,
		bloom:         bloom,
		ledger:        ledger,
		reorgHandler:  reorgHandler,
		confirmations: confirmations,
	}
}

// PreloadAddress 将钱包地址加入布隆过滤器
func (p *Processor) PreloadAddress(address string) {
	if p.bloom == nil {
		return
	}
	p.bloom.Add([]byte(address))
}

// ShouldHandle 判断交易是否关注
func (p *Processor) ShouldHandle(toAddress string) bool {
	if p.bloom == nil {
		return true
	}
	return p.bloom.Test([]byte(toAddress))
}

// HandleEvent 处理扫描到的充值事件
func (p *Processor) HandleEvent(ctx context.Context, event domain.DepositEvent) error {
	if !p.ShouldHandle(event.ToAddress) {
		return nil
	}
	event.Confirmations = max(event.Confirmations, p.confirmations)
	if err := p.consumer.HandleDepositEvent(ctx, event); err != nil {
		return err
	}
	if p.ledger != nil {
		record := domain.DepositRecord{
			TxHash:        event.TxHash,
			Chain:         event.Chain,
			AssetSymbol:   event.AssetSymbol,
			Amount:        event.Amount,
			FromAddress:   event.FromAddress,
			ToAddress:     event.ToAddress,
			Confirmations: event.Confirmations,
			Status:        domain.DepositConfirmed,
			ObservedAt:    timestampOrNow(event.ObservedAt),
		}
		if record.Amount == nil {
			record.Amount = big.NewInt(0)
		}
		if err := p.ledger.AppendDeposit(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

// SetReorgHandler 设置重组处理器（允许运行时替换）
func (p *Processor) SetReorgHandler(handler ReorgHandler) {
	p.reorgHandler = handler
}

// HandleReorg 处理链上重组
func (p *Processor) HandleReorg(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) error {
	if p.reorgHandler == nil {
		return nil
	}
	return p.reorgHandler.OnReorg(ctx, chain, fromHeight, toHeight)
}

func timestampOrNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
