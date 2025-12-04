package service

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
	"github.com/jinmu/go-blockchain/internal/wallet/notification"
	"github.com/jinmu/go-blockchain/internal/wallet/risk"
	"github.com/jinmu/go-blockchain/internal/wallet/scanner"
	"github.com/jinmu/go-blockchain/internal/wallet/signer"
	"github.com/jinmu/go-blockchain/internal/wallet/store"
)

// Manager 集成资产管理、充值监听、提现流程等核心逻辑
type Manager struct {
	assets      store.AssetRepository
	accounts    store.AccountRepository
	balances    store.BalanceRepository
	deposits    store.DepositRepository
	withdrawals store.WithdrawalRepository
	risk        risk.Controller
	signer      signer.Signer
	notifier    *notification.Manager // 通知管理器（可选）

	scannersMu sync.RWMutex
	scanners   map[domain.ChainType]scanner.Scanner
}

// Option 自定义 Manager 行为
type Option func(*Manager)

// WithRiskController 指定风控实现
func WithRiskController(r risk.Controller) Option {
	return func(m *Manager) { m.risk = r }
}

// WithSigner 指定签名机
func WithSigner(s signer.Signer) Option {
	return func(m *Manager) { m.signer = s }
}

// WithNotifier 指定通知管理器
func WithNotifier(n *notification.Manager) Option {
	return func(m *Manager) { m.notifier = n }
}

// NewManager 创建新的钱包管理器
func NewManager(repos store.RepositoryProvider, opts ...Option) *Manager {
	mgr := &Manager{
		assets:      repos,
		accounts:    repos,
		balances:    repos,
		deposits:    repos,
		withdrawals: repos,
		risk:        risk.NoopController{},
		signer:      signer.NoopSigner{},
		scanners:    make(map[domain.ChainType]scanner.Scanner),
	}
	for _, opt := range opts {
		opt(mgr)
	}
	return mgr
}

// RegisterAsset 注册新资产
func (m *Manager) RegisterAsset(ctx context.Context, asset domain.Asset) error {
	if asset.Symbol == "" {
		return fmt.Errorf("wallet: asset symbol required")
	}
	if asset.Decimals == 0 {
		asset.Decimals = 18
	}
	return m.assets.SaveAsset(ctx, asset)
}

// GetAsset 查询资产配置
func (m *Manager) GetAsset(ctx context.Context, symbol string) (domain.Asset, error) {
	return m.assets.GetAsset(ctx, symbol)
}

// EnsureAccount 为用户和资产创建钱包账户
func (m *Manager) EnsureAccount(ctx context.Context, userID string, assetSymbol string) (domain.WalletAccount, error) {
	asset, err := m.assets.GetAsset(ctx, assetSymbol)
	if err != nil {
		return domain.WalletAccount{}, err
	}
	account, err := m.accounts.GetAccount(ctx, userID, assetSymbol)
	if err == nil {
		return account, nil
	}
	address, err := m.signer.GenerateAddress(ctx, asset.Chain, map[string]string{
		"user_id": userID,
		"asset":   assetSymbol,
	})
	if err != nil {
		return domain.WalletAccount{}, err
	}
	account = domain.WalletAccount{
		UserID:      userID,
		AssetSymbol: assetSymbol,
		Address:     address,
		Chain:       asset.Chain,
		CreatedAt:   time.Now(),
	}
	if err := m.accounts.SaveAccount(ctx, account); err != nil {
		return domain.WalletAccount{}, err
	}
	return account, nil
}

// RegisterScanner 注册区块链扫描器
func (m *Manager) RegisterScanner(scanner scanner.Scanner) {
	m.scannersMu.Lock()
	defer m.scannersMu.Unlock()
	m.scanners[scanner.Chain()] = scanner
}

// GetScannerStatuses 获取所有扫描器的状态
func (m *Manager) GetScannerStatuses(ctx context.Context) map[domain.ChainType]scanner.ScannerStatus {
	m.scannersMu.RLock()
	defer m.scannersMu.RUnlock()
	
	statuses := make(map[domain.ChainType]scanner.ScannerStatus)
	for chain, sc := range m.scanners {
		statuses[chain] = sc.GetStatus(ctx)
	}
	return statuses
}

// StartDepositListeners 启动所有扫描器监听充值
func (m *Manager) StartDepositListeners(ctx context.Context) {
	m.scannersMu.RLock()
	defer m.scannersMu.RUnlock()
	for _, sc := range m.scanners {
		scanner := sc
		go func() {
			// 提供重组处理回调（如果需要，可以扩展 Manager 支持自定义重组处理）
			reorgHandler := func(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) error {
				// 默认不做处理，由扫描器内部的重组检测器处理
				// 如果需要在这里处理，可以调用 deposit.Processor.HandleReorg
				return nil
			}
			_ = scanner.Subscribe(ctx, m.HandleDepositEvent, reorgHandler)
		}()
	}
}

// HandleDepositEvent 处理充值事件
func (m *Manager) HandleDepositEvent(ctx context.Context, event domain.DepositEvent) error {
	// 查找账户
	account, err := m.accounts.FindAccountByAddress(ctx, event.ToAddress, event.AssetSymbol)
	if err != nil {
		// 如果找不到账户，记录但不处理（可能是非关注地址）
		return nil
	}

	// 检查是否已存在该充值记录
	existing, err := m.deposits.GetDeposit(ctx, event.TxHash)
	if err == nil {
		// 已存在，更新确认数
		existing.Confirmations = event.Confirmations
		oldStatus := existing.Status
		if existing.Confirmations >= existing.RequiredConfirmations && existing.Status == domain.DepositPending {
			existing.Status = domain.DepositConfirmed
			if err := m.deposits.UpdateDeposit(ctx, existing); err != nil {
				return err
			}
			// 发送确认通知
			if m.notifier != nil && oldStatus != domain.DepositConfirmed {
				go func() {
					if err := m.notifier.NotifyDeposit(context.Background(), existing); err != nil {
						fmt.Printf("Failed to send deposit confirmation notification: %v\n", err)
					}
				}()
			}
			// 如果确认数足够，执行入账
			return m.creditDeposit(ctx, account.UserID, existing)
		}
		return m.deposits.UpdateDeposit(ctx, existing)
	}

	// 创建新的充值记录
	requiredConfirmations := m.getRequiredConfirmations(event.Chain)
	record := domain.DepositRecord{
		TxHash:                event.TxHash,
		UserID:                account.UserID,
		Chain:                 event.Chain,
		AssetSymbol:           event.AssetSymbol,
		Amount:                event.Amount,
		FromAddress:           event.FromAddress,
		ToAddress:             event.ToAddress,
		BlockHeight:           event.BlockHeight,
		Confirmations:         event.Confirmations,
		RequiredConfirmations: requiredConfirmations,
		Status:                domain.DepositPending,
		ObservedAt:            event.ObservedAt,
		Metadata:              event.Metadata,
	}

	// 保存充值记录
	if err := m.deposits.SaveDeposit(ctx, account.UserID, record); err != nil {
		return err
	}

	// 如果确认数足够，立即入账
	if record.Confirmations >= record.RequiredConfirmations {
		record.Status = domain.DepositConfirmed
		if err := m.deposits.UpdateDeposit(ctx, record); err != nil {
			return err
		}
		return m.creditDeposit(ctx, account.UserID, record)
	}

	return nil
}

// getRequiredConfirmations 获取链的所需确认数
func (m *Manager) getRequiredConfirmations(chain domain.ChainType) uint64 {
	switch chain {
	case domain.ChainBitcoin:
		return 6
	case domain.ChainEVM:
		return 12
	case domain.ChainSolana:
		return 32
	default:
		return 6
	}
}

func (m *Manager) creditDeposit(ctx context.Context, userID string, record domain.DepositRecord) error {
	// 检查是否已经入账
	if record.Status == domain.DepositCredited {
		return nil
	}

	// 增加余额
	if err := m.balances.Credit(ctx, userID, record.AssetSymbol, record.Amount); err != nil {
		return err
	}

	// 更新状态
	record.Status = domain.DepositCredited
	record.CreditedAt = time.Now()
	if err := m.deposits.UpdateDeposit(ctx, record); err != nil {
		return err
	}

	// 发送通知（异步，不阻塞）
	if m.notifier != nil {
		go func() {
			if err := m.notifier.NotifyDeposit(context.Background(), record); err != nil {
				// 记录错误但不影响主流程
				fmt.Printf("Failed to send deposit notification: %v\n", err)
			}
		}()
	}

	return nil
}

// ManualCredit 用于人工确认充值到账
func (m *Manager) ManualCredit(ctx context.Context, txHash string) error {
	record, err := m.deposits.GetDeposit(ctx, txHash)
	if err != nil {
		return err
	}
	if record.Status == domain.DepositCredited {
		return nil
	}
	account, err := m.accounts.FindAccountByAddress(ctx, record.ToAddress, record.AssetSymbol)
	if err != nil {
		return err
	}
	return m.creditDeposit(ctx, account.UserID, record)
}

// CreateWithdrawal 发起提现请求
func (m *Manager) CreateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) (domain.WithdrawalRequest, error) {
	if err := m.balances.Freeze(ctx, req.UserID, req.AssetSymbol, req.Amount); err != nil {
		return domain.WithdrawalRequest{}, err
	}
	req.Status = domain.WithdrawalUnderReview
	req.CreatedAt = time.Now()
	req.UpdatedAt = req.CreatedAt
	if err := m.withdrawals.SaveWithdrawal(ctx, req); err != nil {
		_ = m.balances.Unfreeze(ctx, req.UserID, req.AssetSymbol, req.Amount)
		return domain.WithdrawalRequest{}, err
	}
	decision, err := m.risk.EvaluateWithdrawal(ctx, req)
	if err != nil {
		return domain.WithdrawalRequest{}, err
	}
	req.RiskScore = decision.Score
	req.RiskRemarks = decision.Remarks
	if !decision.Approved {
		req.Status = domain.WithdrawalRejected
		_ = m.balances.Unfreeze(ctx, req.UserID, req.AssetSymbol, req.Amount)
		_ = m.withdrawals.UpdateWithdrawal(ctx, req)
		return domain.WithdrawalRequest{}, domain.ErrWithdrawalRejected
	}
	req.Status = domain.WithdrawalApproved
	if err := m.withdrawals.UpdateWithdrawal(ctx, req); err != nil {
		return domain.WithdrawalRequest{}, err
	}
	sig, err := m.signer.SignWithdrawal(ctx, req)
	if err != nil {
		req.Status = domain.WithdrawalFailed
		_ = m.withdrawals.UpdateWithdrawal(ctx, req)
		_ = m.balances.Unfreeze(ctx, req.UserID, req.AssetSymbol, req.Amount)
		return domain.WithdrawalRequest{}, domain.ErrSignerUnavailable
	}
	req.Status = domain.WithdrawalSigned
	req.RawTx = sig.RawTx
	req.TxHash = sig.TxHash
	req.Metadata = mergeStringMap(req.Metadata, sig.Metadata)
	if err := m.withdrawals.UpdateWithdrawal(ctx, req); err != nil {
		return domain.WithdrawalRequest{}, err
	}
	if err := m.commitWithdrawal(ctx, req); err != nil {
		return domain.WithdrawalRequest{}, err
	}
	return req, nil
}

func (m *Manager) commitWithdrawal(ctx context.Context, req domain.WithdrawalRequest) error {
	if err := m.balances.Debit(ctx, req.UserID, req.AssetSymbol, req.Amount); err != nil {
		return err
	}
	if err := m.balances.Unfreeze(ctx, req.UserID, req.AssetSymbol, req.Amount); err != nil {
		return err
	}
	req.Status = domain.WithdrawalCompleted
	return m.withdrawals.UpdateWithdrawal(ctx, req)
}

// GetBalance 查询余额
func (m *Manager) GetBalance(ctx context.Context, userID, asset string) (domain.Balance, error) {
	return m.balances.GetBalance(ctx, userID, asset)
}

// TransferBetweenAccounts 用于资金调度
func (m *Manager) TransferBetweenAccounts(ctx context.Context, fromUser, toUser, asset string, amount *big.Int) error {
	if err := m.balances.Debit(ctx, fromUser, asset, amount); err != nil {
		return err
	}
	if err := m.balances.Credit(ctx, toUser, asset, amount); err != nil {
		return err
	}
	return nil
}

// GetAccount 获取账户信息
func (m *Manager) GetAccount(ctx context.Context, userID, asset string) (domain.WalletAccount, error) {
	return m.accounts.GetAccount(ctx, userID, asset)
}

// GetDeposit 获取充值记录
func (m *Manager) GetDeposit(ctx context.Context, txHash string) (domain.DepositRecord, error) {
	return m.deposits.GetDeposit(ctx, txHash)
}

// ListDeposits 查询充值记录
func (m *Manager) ListDeposits(ctx context.Context, userID, assetSymbol string, status domain.DepositStatus, limit, offset int) ([]domain.DepositRecord, error) {
	return m.deposits.ListDeposits(ctx, userID, assetSymbol, status, limit, offset)
}

// GetDepositStatistics 获取充值统计信息
func (m *Manager) GetDepositStatistics(ctx context.Context, startTime, endTime *time.Time) (store.DepositStatistics, error) {
	return m.deposits.GetDepositStatistics(ctx, startTime, endTime)
}

// GetWithdrawal 获取提现记录
func (m *Manager) GetWithdrawal(ctx context.Context, id string) (domain.WithdrawalRequest, error) {
	return m.withdrawals.GetWithdrawal(ctx, id)
}

// GetRiskController 获取风控控制器
func (m *Manager) GetRiskController() risk.Controller {
	return m.risk
}

// ListAssets 列出所有资产
func (m *Manager) ListAssets(ctx context.Context) ([]domain.Asset, error) {
	return m.assets.ListAssets(ctx)
}

// GetBalanceRepository 获取余额仓储（用于管理员操作）
func (m *Manager) GetBalanceRepository() store.BalanceRepository {
	return m.balances
}

// RollbackDeposit 回滚单笔充值（实现 deposit.ReorgManager 接口）
func (m *Manager) RollbackDeposit(ctx context.Context, record domain.DepositRecord) error {
	// 查找对应的用户账户
	account, err := m.accounts.FindAccountByAddress(ctx, record.ToAddress, record.AssetSymbol)
	if err != nil {
		return err
	}

	// 如果已经入账，需要从余额中扣除
	if record.Status == domain.DepositCredited {
		if err := m.balances.Debit(ctx, account.UserID, record.AssetSymbol, record.Amount); err != nil {
			return err
		}
	}

	// 更新充值记录状态
	record.Status = domain.DepositFailed
	return m.deposits.UpdateDeposit(ctx, record)
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
