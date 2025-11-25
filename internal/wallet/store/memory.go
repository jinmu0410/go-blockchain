package store

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// InMemoryStore 提供简单的内存实现, 方便开发和单元测试
type InMemoryStore struct {
	mu           sync.RWMutex
	assets       map[string]domain.Asset
	accounts     map[string]domain.WalletAccount
	balances     map[string]map[string]*domain.Balance
	deposits     map[string]domain.DepositRecord
	withdrawals  map[string]domain.WithdrawalRequest
	addresses    map[string]string           // address -> accountKey
	blocks       map[string]BlockInfo        // chain:height -> BlockInfo
	latestBlocks map[domain.ChainType]uint64 // chain -> latest height
}

// NewInMemoryStore 创建内存仓储
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		assets:       make(map[string]domain.Asset),
		accounts:     make(map[string]domain.WalletAccount),
		balances:     make(map[string]map[string]*domain.Balance),
		deposits:     make(map[string]domain.DepositRecord),
		withdrawals:  make(map[string]domain.WithdrawalRequest),
		addresses:    make(map[string]string),
		blocks:       make(map[string]BlockInfo),
		latestBlocks: make(map[domain.ChainType]uint64),
	}
}

func (s *InMemoryStore) assetKey(symbol string) string {
	return symbol
}

func (s *InMemoryStore) accountKey(userID, asset string) string {
	return userID + ":" + asset
}

// SaveAsset implements AssetRepository
func (s *InMemoryStore) SaveAsset(ctx context.Context, asset domain.Asset) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.assets[s.assetKey(asset.Symbol)]; ok {
		return domain.ErrAssetExists
	}
	s.assets[s.assetKey(asset.Symbol)] = asset
	return nil
}

// GetAsset implements AssetRepository
func (s *InMemoryStore) GetAsset(ctx context.Context, symbol string) (domain.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	asset, ok := s.assets[s.assetKey(symbol)]
	if !ok {
		return domain.Asset{}, domain.ErrAssetNotFound
	}
	return asset, nil
}

// ListAssets implements AssetRepository
func (s *InMemoryStore) ListAssets(ctx context.Context) ([]domain.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]domain.Asset, 0, len(s.assets))
	for _, asset := range s.assets {
		list = append(list, asset)
	}
	return list, nil
}

// SaveAccount implements AccountRepository
func (s *InMemoryStore) SaveAccount(ctx context.Context, account domain.WalletAccount) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.accountKey(account.UserID, account.AssetSymbol)
	if _, ok := s.accounts[key]; ok {
		return domain.ErrAccountExists
	}
	s.accounts[key] = account
	s.addresses[account.Address+":"+account.AssetSymbol] = key
	if _, ok := s.balances[account.UserID]; !ok {
		s.balances[account.UserID] = make(map[string]*domain.Balance)
	}
	if _, ok := s.balances[account.UserID][account.AssetSymbol]; !ok {
		s.balances[account.UserID][account.AssetSymbol] = &domain.Balance{
			Available: big.NewInt(0),
			Frozen:    big.NewInt(0),
			Pending:   big.NewInt(0),
		}
	}
	return nil
}

// GetAccount implements AccountRepository
func (s *InMemoryStore) GetAccount(ctx context.Context, userID, asset string) (domain.WalletAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := s.accountKey(userID, asset)
	account, ok := s.accounts[key]
	if !ok {
		return domain.WalletAccount{}, domain.ErrAccountNotFound
	}
	return account, nil
}

// FindAccountByAddress implements AccountRepository
func (s *InMemoryStore) FindAccountByAddress(ctx context.Context, address, asset string) (domain.WalletAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key, ok := s.addresses[address+":"+asset]
	if !ok {
		return domain.WalletAccount{}, domain.ErrAccountNotFound
	}
	account, ok := s.accounts[key]
	if !ok {
		return domain.WalletAccount{}, domain.ErrAccountNotFound
	}
	return account, nil
}

func cloneBigInt(src *big.Int) *big.Int {
	if src == nil {
		return nil
	}
	return new(big.Int).Set(src)
}

func (s *InMemoryStore) getBalanceLocked(userID, asset string) (*domain.Balance, error) {
	acctKey := s.accountKey(userID, asset)
	if _, ok := s.accounts[acctKey]; !ok {
		return nil, domain.ErrAccountNotFound
	}
	if _, ok := s.balances[userID]; !ok {
		s.balances[userID] = make(map[string]*domain.Balance)
	}
	balance, ok := s.balances[userID][asset]
	if !ok {
		balance = &domain.Balance{Available: big.NewInt(0), Frozen: big.NewInt(0), Pending: big.NewInt(0)}
		s.balances[userID][asset] = balance
	}
	return balance, nil
}

// Credit implements BalanceRepository
func (s *InMemoryStore) Credit(ctx context.Context, userID, asset string, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	balance, err := s.getBalanceLocked(userID, asset)
	if err != nil {
		return err
	}
	balance.Available = new(big.Int).Add(balance.Available, amount)
	return nil
}

// Debit implements BalanceRepository
func (s *InMemoryStore) Debit(ctx context.Context, userID, asset string, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	balance, err := s.getBalanceLocked(userID, asset)
	if err != nil {
		return err
	}
	if balance.Available.Cmp(amount) < 0 {
		return domain.ErrInsufficientFunds
	}
	balance.Available = new(big.Int).Sub(balance.Available, amount)
	return nil
}

// Freeze implements BalanceRepository
func (s *InMemoryStore) Freeze(ctx context.Context, userID, asset string, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	balance, err := s.getBalanceLocked(userID, asset)
	if err != nil {
		return err
	}
	if balance.Available.Cmp(amount) < 0 {
		return domain.ErrInsufficientFunds
	}
	balance.Available = new(big.Int).Sub(balance.Available, amount)
	balance.Frozen = new(big.Int).Add(balance.Frozen, amount)
	return nil
}

// Unfreeze implements BalanceRepository
func (s *InMemoryStore) Unfreeze(ctx context.Context, userID, asset string, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	balance, err := s.getBalanceLocked(userID, asset)
	if err != nil {
		return err
	}
	if balance.Frozen.Cmp(amount) < 0 {
		return domain.ErrInsufficientFunds
	}
	balance.Frozen = new(big.Int).Sub(balance.Frozen, amount)
	balance.Available = new(big.Int).Add(balance.Available, amount)
	return nil
}

// GetBalance implements BalanceRepository
func (s *InMemoryStore) GetBalance(ctx context.Context, userID, asset string) (domain.Balance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	balance, err := s.getBalanceLocked(userID, asset)
	if err != nil {
		return domain.Balance{}, err
	}
	return domain.Balance{
		Available: cloneBigInt(balance.Available),
		Frozen:    cloneBigInt(balance.Frozen),
		Pending:   cloneBigInt(balance.Pending),
	}, nil
}

// SaveDeposit implements DepositRepository
func (s *InMemoryStore) SaveDeposit(ctx context.Context, userID string, record domain.DepositRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if record.ObservedAt.IsZero() {
		record.ObservedAt = time.Now()
	}
	if _, ok := s.deposits[record.TxHash]; ok {
		return nil
	}
	s.deposits[record.TxHash] = record
	return nil
}

// GetDeposit implements DepositRepository
func (s *InMemoryStore) GetDeposit(ctx context.Context, txHash string) (domain.DepositRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.deposits[txHash]
	if !ok {
		return domain.DepositRecord{}, domain.ErrDepositNotFound
	}
	return record, nil
}

// UpdateDeposit implements DepositRepository
func (s *InMemoryStore) UpdateDeposit(ctx context.Context, record domain.DepositRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.deposits[record.TxHash]; !ok {
		return domain.ErrDepositNotFound
	}
	s.deposits[record.TxHash] = record
	return nil
}

// SaveWithdrawal implements WithdrawalRepository
func (s *InMemoryStore) SaveWithdrawal(ctx context.Context, req domain.WithdrawalRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.withdrawals[req.ID]; ok {
		return domain.ErrWithdrawalPending
	}
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}
	req.UpdatedAt = req.CreatedAt
	s.withdrawals[req.ID] = req
	return nil
}

// UpdateWithdrawal implements WithdrawalRepository
func (s *InMemoryStore) UpdateWithdrawal(ctx context.Context, req domain.WithdrawalRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.withdrawals[req.ID]; !ok {
		return domain.ErrWithdrawalNotFound
	}
	req.UpdatedAt = time.Now()
	s.withdrawals[req.ID] = req
	return nil
}

// GetWithdrawal implements WithdrawalRepository
func (s *InMemoryStore) GetWithdrawal(ctx context.Context, id string) (domain.WithdrawalRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.withdrawals[id]
	if !ok {
		return domain.WithdrawalRequest{}, domain.ErrWithdrawalNotFound
	}
	return req, nil
}

// FindDepositsByBlockRange implements DepositRepository
func (s *InMemoryStore) FindDepositsByBlockRange(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) ([]domain.DepositRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	results := make([]domain.DepositRecord, 0)
	for _, record := range s.deposits {
		// 只查找指定链的记录
		if record.Chain != chain {
			continue
		}
		// 检查区块高度是否在范围内
		if record.BlockHeight >= fromHeight && record.BlockHeight <= toHeight {
			results = append(results, record)
		}
	}
	return results, nil
}

func (s *InMemoryStore) blockKey(chain domain.ChainType, height uint64) string {
	return string(chain) + ":" + fmt.Sprintf("%d", height)
}

// SaveBlock implements BlockRepository
func (s *InMemoryStore) SaveBlock(ctx context.Context, chain domain.ChainType, height uint64, hash string, parentHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := s.blockKey(chain, height)
	s.blocks[key] = BlockInfo{
		Chain:      chain,
		Height:     height,
		Hash:       hash,
		ParentHash: parentHash,
		CreatedAt:  time.Now(),
	}
	// 更新最新区块高度
	if height > s.latestBlocks[chain] {
		s.latestBlocks[chain] = height
	}
	return nil
}

// GetBlock implements BlockRepository
func (s *InMemoryStore) GetBlock(ctx context.Context, chain domain.ChainType, height uint64) (BlockInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := s.blockKey(chain, height)
	block, ok := s.blocks[key]
	if !ok {
		return BlockInfo{}, domain.ErrBlockNotFound
	}
	return block, nil
}

// GetLatestBlock implements BlockRepository
func (s *InMemoryStore) GetLatestBlock(ctx context.Context, chain domain.ChainType) (BlockInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	latestHeight, ok := s.latestBlocks[chain]
	if !ok || latestHeight == 0 {
		return BlockInfo{}, domain.ErrBlockNotFound
	}
	key := s.blockKey(chain, latestHeight)
	block, ok := s.blocks[key]
	if !ok {
		return BlockInfo{}, domain.ErrBlockNotFound
	}
	return block, nil
}

// DeleteBlocksFromHeight implements BlockRepository
func (s *InMemoryStore) DeleteBlocksFromHeight(ctx context.Context, chain domain.ChainType, fromHeight uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 删除从 fromHeight 开始的所有区块
	latestHeight := s.latestBlocks[chain]
	for height := fromHeight; height <= latestHeight; height++ {
		key := s.blockKey(chain, height)
		delete(s.blocks, key)
	}
	// 更新最新区块高度
	if fromHeight > 0 {
		s.latestBlocks[chain] = fromHeight - 1
	} else {
		delete(s.latestBlocks, chain)
	}
	return nil
}
