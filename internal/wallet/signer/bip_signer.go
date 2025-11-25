package signer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jinmu/go-blockchain/internal/wallet/bip"
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// BIPSigner 基于 BIP 协议的签名机实现
type BIPSigner struct {
	generator *bip.SimpleBIPGenerator
	keystore  bip.KeyStore
	rpcClient interface{} // RPC 客户端（用于获取 nonce、gas 等）
}

// NewBIPSigner 创建 BIP 签名机
func NewBIPSigner(masterSeed []byte, keystore bip.KeyStore) *BIPSigner {
	return &BIPSigner{
		generator: bip.NewSimpleBIPGenerator(masterSeed),
		keystore:  keystore,
	}
}

// GenerateAddress 生成地址（实现 Signer 接口）
func (s *BIPSigner) GenerateAddress(ctx context.Context, chain domain.ChainType, metadata map[string]string) (string, error) {
	// 从 metadata 获取用户ID和资产
	userID := metadata["user_id"]
	asset := metadata["asset"]

	if userID == "" || asset == "" {
		return "", fmt.Errorf("user_id and asset are required in metadata")
	}

	// 生成地址索引（基于用户ID和资产，确保相同用户+资产总是得到相同地址）
	addressIndex := bip.GenerateAddressIndex(userID, asset)

	// 生成地址
	addressInfo, err := s.generator.GenerateAddress(ctx, chain, 0, addressIndex, metadata)
	if err != nil {
		return "", fmt.Errorf("failed to generate address: %w", err)
	}

	// 加密并保存私钥
	encryptedKey, err := bip.EncryptPrivateKey(addressInfo.PrivateKey, "master_password") // 实际应该使用配置的密码
	if err != nil {
		return "", fmt.Errorf("failed to encrypt private key: %w", err)
	}

	if err := s.keystore.SaveKey(ctx, addressInfo.Address, encryptedKey); err != nil {
		return "", fmt.Errorf("failed to save private key: %w", err)
	}

	return addressInfo.Address, nil
}

// SignWithdrawal 签名提现交易（实现 Signer 接口）
func (s *BIPSigner) SignWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (Result, error) {
	// 1. 从 KeyStore 获取私钥
	encryptedKey, err := s.keystore.GetKey(ctx, req.ToAddress) // 注意：这里应该是热钱包地址，不是用户地址
	if err != nil {
		return Result{}, fmt.Errorf("failed to get private key: %w", err)
	}

	// 2. 解密私钥
	privateKeyBytes, err := bip.DecryptPrivateKey(encryptedKey, "master_password") // 实际应该使用配置的密码
	if err != nil {
		return Result{}, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// 3. 转换为 ECDSA 私钥
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return Result{}, fmt.Errorf("failed to convert private key: %w", err)
	}

	// 4. 构建交易（EVM 链示例）
	if req.Chain == domain.ChainEVM {
		return s.signEVMTransaction(ctx, req, privateKey)
	}

	return Result{}, fmt.Errorf("unsupported chain: %s", req.Chain)
}

// signEVMTransaction 签名 EVM 交易
func (s *BIPSigner) signEVMTransaction(ctx context.Context, req domain.WithdrawalRequest, privateKey *ecdsa.PrivateKey) (Result, error) {
	// 从 metadata 获取 nonce 和 gas 信息
	nonceStr := req.Metadata["nonce"]
	gasPriceStr := req.Metadata["gas_price"]
	gasLimitStr := req.Metadata["gas_limit"]

	nonce := uint64(0)
	if nonceStr != "" {
		var ok bool
		nonceInt, ok := new(big.Int).SetString(nonceStr, 10)
		if !ok {
			return Result{}, fmt.Errorf("invalid nonce: %s", nonceStr)
		}
		nonce = nonceInt.Uint64()
	}

	gasPrice := big.NewInt(0)
	if gasPriceStr != "" {
		var ok bool
		gasPrice, ok = new(big.Int).SetString(gasPriceStr, 10)
		if !ok {
			return Result{}, fmt.Errorf("invalid gas price: %s", gasPriceStr)
		}
	}

	gasLimit := uint64(21000) // 默认值
	if gasLimitStr != "" {
		var ok bool
		gasLimitInt, ok := new(big.Int).SetString(gasLimitStr, 10)
		if ok {
			gasLimit = gasLimitInt.Uint64()
		}
	}

	// 获取链ID（从 metadata 或配置）
	chainID := big.NewInt(1) // 默认 Ethereum 主网
	if chainIDStr := req.Metadata["chain_id"]; chainIDStr != "" {
		var ok bool
		chainID, ok = new(big.Int).SetString(chainIDStr, 10)
		if !ok {
			return Result{}, fmt.Errorf("invalid chain ID: %s", chainIDStr)
		}
	}

	// 构建交易
	toAddress := common.HexToAddress(req.ToAddress)
	var tx *types.Transaction

	// 检查是否使用 EIP-1559
	if req.Metadata["max_fee_per_gas"] != "" {
		// EIP-1559 交易
		maxFeePerGas, _ := new(big.Int).SetString(req.Metadata["max_fee_per_gas"], 10)
		maxPriorityFeePerGas, _ := new(big.Int).SetString(req.Metadata["max_priority_fee_per_gas"], 10)

		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			GasTipCap: maxPriorityFeePerGas,
			GasFeeCap: maxFeePerGas,
			Gas:       gasLimit,
			To:        &toAddress,
			Value:     req.Amount,
			Data:      nil,
		})
	} else {
		// Legacy 交易
		tx = types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			To:       &toAddress,
			Value:    req.Amount,
			Data:     nil,
		})
	}

	// 签名交易
	signer := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return Result{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 序列化交易
	txBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return Result{}, fmt.Errorf("failed to marshal transaction: %w", err)
	}

	return Result{
		RawTx:  txBytes,
		TxHash: signedTx.Hash().Hex(),
		Metadata: map[string]string{
			"nonce":      fmt.Sprintf("%d", nonce),
			"gas_price":  gasPrice.String(),
			"gas_limit":  fmt.Sprintf("%d", gasLimit),
			"chain_id":   chainID.String(),
		},
	}, nil
}

