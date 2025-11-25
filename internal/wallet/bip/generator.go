package bip

import (
	"context"
	"fmt"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// Generator BIP 钱包生成器接口
type Generator interface {
	// GenerateMnemonic 生成助记词（BIP39）
	GenerateMnemonic(strength int) (string, error)

	// GenerateWalletFromMnemonic 从助记词生成钱包
	GenerateWalletFromMnemonic(mnemonic string, password string) (*WalletInfo, error)

	// DeriveKeyPair 派生密钥对（BIP32/BIP44）
	DeriveKeyPair(wallet *WalletInfo, path DerivationPath) (*KeyPair, error)

	// DeriveAddress 派生地址（根据链类型）
	DeriveAddress(chain domain.ChainType, keyPair *KeyPair) (string, error)

	// GenerateAddress 生成地址（完整流程）
	GenerateAddress(ctx context.Context, chain domain.ChainType, accountIndex, addressIndex uint32, metadata map[string]string) (*AddressInfo, error)
}

// BIPGenerator BIP 钱包生成器实现
type BIPGenerator struct {
	masterMnemonic string // 主助记词（用于生成所有地址）
	masterPassword string // 主密码（用于加密种子）
}

// NewBIPGenerator 创建 BIP 生成器
func NewBIPGenerator(masterMnemonic, masterPassword string) *BIPGenerator {
	return &BIPGenerator{
		masterMnemonic: masterMnemonic,
		masterPassword: masterPassword,
	}
}

// GenerateMnemonic 生成助记词（BIP39）
// strength: 128 (12 words), 256 (24 words)
func (g *BIPGenerator) GenerateMnemonic(strength int) (string, error) {
	// TODO: 使用 bip39 库生成助记词
	// 示例实现（需要引入 bip39 库）
	// entropy, err := bip39.NewEntropy(strength)
	// if err != nil {
	//     return "", err
	// }
	// mnemonic, err := bip39.NewMnemonic(entropy)
	// return mnemonic, err

	// 临时实现：返回错误提示需要实现
	return "", fmt.Errorf("bip39 mnemonic generation not implemented, please use external library")
}

// GenerateWalletFromMnemonic 从助记词生成钱包
func (g *BIPGenerator) GenerateWalletFromMnemonic(mnemonic string, password string) (*WalletInfo, error) {
	// TODO: 使用 bip39 和 bip32 库
	// 1. 从助记词生成种子: seed = PBKDF2(mnemonic, password)
	// 2. 从种子生成主密钥: masterKey, chainCode = HMAC-SHA512(seed, "Bitcoin seed")

	return nil, fmt.Errorf("wallet generation from mnemonic not implemented, please use external library")
}

// DeriveKeyPair 派生密钥对（BIP32/BIP44）
func (g *BIPGenerator) DeriveKeyPair(wallet *WalletInfo, path DerivationPath) (*KeyPair, error) {
	// TODO: 使用 bip32 库派生
	// 1. 从主密钥派生到指定路径
	// 2. 生成私钥和公钥
	// 3. 根据链类型生成地址

	return nil, fmt.Errorf("key derivation not implemented, please use external library")
}

// DeriveAddress 派生地址（根据链类型）
func (g *BIPGenerator) DeriveAddress(chain domain.ChainType, keyPair *KeyPair) (string, error) {
	simpleGen := &SimpleBIPGenerator{}
	return simpleGen.deriveAddress(chain, keyPair)
}

// deriveEVMAddress 派生 EVM 地址（已移动到 implementation.go）
// deriveBitcoinAddress 派生 Bitcoin 地址（已移动到 implementation.go）
// deriveSolanaAddress 派生 Solana 地址（已移动到 implementation.go）

// GenerateAddress 生成地址（完整流程）
// accountIndex: 账户索引（通常为 0）
// addressIndex: 地址索引（每个用户/资产使用不同的索引）
func (g *BIPGenerator) GenerateAddress(ctx context.Context, chain domain.ChainType, accountIndex, addressIndex uint32, metadata map[string]string) (*AddressInfo, error) {
	// 使用 SimpleBIPGenerator 实现
	// 注意：实际应该从助记词生成种子
	simpleGen := &SimpleBIPGenerator{
		masterSeed: []byte("default-master-seed-32-bytes-need-to-be-replaced"),
	}
	return simpleGen.GenerateAddress(ctx, chain, accountIndex, addressIndex, metadata)
}

// BuildDerivationPath 构建派生路径
func BuildDerivationPath(chain domain.ChainType, accountIndex, addressIndex uint32) DerivationPath {
	coinType := GetCoinType(chain)
	return DerivationPath{
		Purpose:      44,
		CoinType:     coinType,
		Account:      accountIndex,
		Change:       0,
		AddressIndex: addressIndex,
	}
}

// FormatDerivationPath 格式化派生路径为字符串
func FormatDerivationPath(path DerivationPath) string {
	return fmt.Sprintf("m/%d'/%d'/%d'/%d/%d",
		path.Purpose,
		path.CoinType,
		path.Account,
		path.Change,
		path.AddressIndex,
	)
}
