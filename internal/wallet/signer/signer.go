package signer

import (
	"context"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// Result 签名结果
type Result struct {
	RawTx    []byte
	TxHash   string
	Metadata map[string]string
}

// Signer 定义签名机接口
type Signer interface {
	GenerateAddress(ctx context.Context, chain domain.ChainType, metadata map[string]string) (string, error)
	SignWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (Result, error)
}

// NoopSigner 提供伪实现
type NoopSigner struct{}

// GenerateAddress 返回伪地址
func (NoopSigner) GenerateAddress(ctx context.Context, chain domain.ChainType, metadata map[string]string) (string, error) {
	return "0x0000000000000000000000000000000000000000", nil
}

// SignWithdrawal 返回空签名
func (NoopSigner) SignWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (Result, error) {
	return Result{RawTx: []byte("signed"), TxHash: "0xmock"}, nil
}
