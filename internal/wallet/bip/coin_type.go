package bip

import (
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// CoinType BIP44 币种类型
// 参考: https://github.com/satoshilabs/slips/blob/master/slip-0044.md
const (
	CoinTypeBitcoin  uint32 = 0
	CoinTypeEthereum uint32 = 60
	CoinTypeSolana   uint32 = 501
	CoinTypePolygon  uint32 = 966
	CoinTypeBSC      uint32 = 9006
)

// GetCoinType 根据链类型获取币种类型
func GetCoinType(chain domain.ChainType) uint32 {
	switch chain {
	case domain.ChainBitcoin:
		return CoinTypeBitcoin
	case domain.ChainEVM:
		return CoinTypeEthereum // 默认使用 Ethereum，可以根据具体链调整
	case domain.ChainSolana:
		return CoinTypeSolana
	default:
		return CoinTypeEthereum
	}
}

// GetCoinTypeBySymbol 根据资产符号获取币种类型
func GetCoinTypeBySymbol(symbol string) uint32 {
	switch symbol {
	case "BTC":
		return CoinTypeBitcoin
	case "ETH", "USDT", "USDC": // ERC20 代币使用 Ethereum 路径
		return CoinTypeEthereum
	case "MATIC":
		return CoinTypePolygon
	case "BNB":
		return CoinTypeBSC
	case "SOL":
		return CoinTypeSolana
	default:
		return CoinTypeEthereum
	}
}

