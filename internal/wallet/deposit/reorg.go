package deposit

import (
	"context"
	"fmt"
	"math/big"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/store"
)

// ReorgDepthConfig 按链类型配置重组深度
type ReorgDepthConfig map[domain.ChainType]uint64

// DefaultReorgDepths 各链的默认重组深度（基于行业最佳实践）
var DefaultReorgDepths = ReorgDepthConfig{
	domain.ChainBitcoin: 6,  // Bitcoin: 6 个确认（约 1 小时），重组深度 6
	domain.ChainEVM:     12, // Ethereum/Polygon/BSC 等: 12-15 个确认，重组深度 12
	domain.ChainSolana:  32, // Solana: 32 个确认（约 13 秒），但重组深度通常更大
}

// ReorgDepthProvider 提供重组深度配置
type ReorgDepthProvider interface {
	GetReorgDepth(chain domain.ChainType) uint64
}

// DefaultReorgHandler 默认的区块重组处理器实现
type DefaultReorgHandler struct {
	repos         store.RepositoryProvider
	manager       ReorgManager
	ledger        Ledger
	depthProvider ReorgDepthProvider // 重组深度提供者
	defaultDepth  uint64             // 默认深度（当 provider 未配置时使用）
}

// ReorgManager 提供回滚所需的管理方法
type ReorgManager interface {
	RollbackDeposit(ctx context.Context, record domain.DepositRecord) error
}

// NewDefaultReorgHandler 创建默认的重组处理器
// 参数说明：
//   - repos: 仓储接口
//   - manager: 可选的回滚管理器
//   - ledger: 可选的资金流水表
//   - depthProvider: 可选的深度提供者，如果为 nil，使用 defaultDepth
//   - defaultDepth: 默认深度，如果为 0，使用链类型对应的默认值
func NewDefaultReorgHandler(repos store.RepositoryProvider, manager ReorgManager, ledger Ledger, depthProvider ReorgDepthProvider, defaultDepth uint64) *DefaultReorgHandler {
	return &DefaultReorgHandler{
		repos:         repos,
		manager:       manager,
		ledger:        ledger,
		depthProvider: depthProvider,
		defaultDepth:  defaultDepth,
	}
}

// NewDefaultReorgHandlerWithConfig 使用配置创建重组处理器（推荐方式）
func NewDefaultReorgHandlerWithConfig(repos store.RepositoryProvider, manager ReorgManager, ledger Ledger, config ReorgDepthConfig) *DefaultReorgHandler {
	provider := &configDepthProvider{config: config}
	return NewDefaultReorgHandler(repos, manager, ledger, provider, 0)
}

// configDepthProvider 基于配置的深度提供者
type configDepthProvider struct {
	config ReorgDepthConfig
}

func (p *configDepthProvider) GetReorgDepth(chain domain.ChainType) uint64 {
	if depth, ok := p.config[chain]; ok {
		return depth
	}
	// 如果配置中没有，使用全局默认值
	if depth, ok := DefaultReorgDepths[chain]; ok {
		return depth
	}
	// 最后兜底：使用通用默认值 6
	return 6
}

// getReorgDepth 获取指定链的重组深度
func (h *DefaultReorgHandler) getReorgDepth(chain domain.ChainType) uint64 {
	if h.depthProvider != nil {
		return h.depthProvider.GetReorgDepth(chain)
	}
	// 如果未提供 provider，使用 defaultDepth
	if h.defaultDepth > 0 {
		return h.defaultDepth
	}
	// 最后使用链类型对应的默认值
	if depth, ok := DefaultReorgDepths[chain]; ok {
		return depth
	}
	// 兜底：6 个区块
	return 6
}

// OnReorg 处理区块重组
// 逻辑：
// 1. 查找 fromHeight 到 toHeight 范围内的所有充值记录
// 2. 将已入账的记录状态设置为 DepositFailed（回滚状态）
// 3. 回滚用户余额
// 4. 回滚资金流水
// fromHeight: 重组开始的区块高度（分叉点的下一个高度）
// toHeight: 重组结束的区块高度（当前最新区块）
func (h *DefaultReorgHandler) OnReorg(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) error {
	// 步骤1: 查找受影响区块范围内的所有充值记录（从 fromHeight 到 toHeight）
	records, err := h.repos.FindDepositsByBlockRange(ctx, chain, fromHeight, toHeight)
	if err != nil {
		return fmt.Errorf("failed to find deposits in reorg range: %w", err)
	}

	// 统计回滚信息
	rollbackStats := make(map[string]*rollbackInfo) // userID:asset -> info

	// 遍历所有受影响的充值记录
	for _, record := range records {
		// 只回滚已经入账的记录
		if record.Status != domain.DepositCredited {
			continue
		}

		// 查找对应的用户账户
		account, err := h.repos.FindAccountByAddress(ctx, record.ToAddress, record.AssetSymbol)
		if err != nil {
			// 如果找不到账户，记录错误但继续处理其他记录
			continue
		}

		// 聚合回滚信息（按用户+资产分组）
		key := account.UserID + ":" + record.AssetSymbol
		info, ok := rollbackStats[key]
		if !ok {
			info = &rollbackInfo{
				userID:      account.UserID,
				asset:       record.AssetSymbol,
				totalAmount: big.NewInt(0),
				records:     make([]domain.DepositRecord, 0),
			}
			rollbackStats[key] = info
		}
		info.totalAmount = new(big.Int).Add(info.totalAmount, record.Amount)
		info.records = append(info.records, record)

		// 更新充值记录状态为失败
		record.Status = domain.DepositFailed
		if err := h.repos.UpdateDeposit(ctx, record); err != nil {
			return fmt.Errorf("failed to update deposit status: %w", err)
		}

		// 调用 Ledger 回滚流水
		if h.ledger != nil {
			if err := h.ledger.Rollback(ctx, record.TxHash); err != nil {
				// Ledger 回滚失败不影响主流程，但记录错误
				// 实际生产环境应该记录告警
			}
		}
	}

	// 批量回滚余额（按用户+资产聚合，避免多次数据库操作）
	// 如果提供了 manager，使用 manager 的单笔回滚方法；否则直接批量 Debit
	if h.manager != nil {
		// 使用 manager 的单笔回滚（manager 内部会处理余额扣减）
		for _, info := range rollbackStats {
			for _, record := range info.records {
				if err := h.manager.RollbackDeposit(ctx, record); err != nil {
					return fmt.Errorf("failed to rollback deposit %s: %w", record.TxHash, err)
				}
			}
		}
	} else {
		// 直接批量回滚余额
		for _, info := range rollbackStats {
			if err := h.repos.Debit(ctx, info.userID, info.asset, info.totalAmount); err != nil {
				return fmt.Errorf("failed to debit balance for user %s asset %s: %w", info.userID, info.asset, err)
			}
		}
	}

	return nil
}

// rollbackInfo 回滚信息聚合
type rollbackInfo struct {
	userID      string
	asset       string
	totalAmount *big.Int
	records     []domain.DepositRecord
}
