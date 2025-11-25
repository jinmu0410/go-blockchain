package app

import (
	"context"
	"fmt"
	"log"

	"github.com/jinmu/go-blockchain/internal/config"
	"github.com/jinmu/go-blockchain/internal/wallet/rpc"
	"github.com/jinmu/go-blockchain/internal/wallet/risk"
	"github.com/jinmu/go-blockchain/internal/wallet/service"
	"github.com/jinmu/go-blockchain/internal/wallet/signer"
	"github.com/jinmu/go-blockchain/internal/wallet/store"
	"github.com/jinmu/go-blockchain/internal/wallet/domain"
)

// App 应用结构
type App struct {
	Config   *config.Config
	Manager  *service.Manager
	Store    store.RepositoryProvider
	Signer   signer.Signer
	RPC      map[domain.ChainType]rpc.Client
}

// NewApp 创建应用实例
func NewApp(cfg *config.Config) (*App, error) {
	// 1. 初始化存储
	store := store.NewInMemoryStore()

	// 2. 初始化签名机
	masterSeed := []byte(cfg.Wallet.MasterSeed)
	if len(masterSeed) == 0 {
		// 如果没有配置，使用默认种子（仅用于开发测试）
		masterSeed = []byte("default-master-seed-32-bytes-for-development-only")
		log.Println("WARNING: Using default master seed, not for production!")
	}
	
	keystore := NewMemoryKeyStore() // 简单的内存 KeyStore 实现
	signer := signer.NewBIPSigner(masterSeed, keystore)

	// 3. 初始化 RPC 客户端
	rpcClients := make(map[domain.ChainType]rpc.Client)
	
	if cfg.RPC.Ethereum != "" {
		evmClient, err := rpc.NewClient(domain.ChainEVM, cfg.RPC.Ethereum)
		if err != nil {
			return nil, fmt.Errorf("failed to create EVM client: %w", err)
		}
		rpcClients[domain.ChainEVM] = evmClient
	}

	// 4. 初始化风控
	riskController := &risk.NoopController{}

	// 5. 初始化 Manager
	manager := service.NewManager(store,
		service.WithRiskController(riskController),
		service.WithSigner(signer),
	)

	// 6. 注册默认资产
	if err := registerDefaultAssets(context.Background(), manager); err != nil {
		return nil, fmt.Errorf("failed to register default assets: %w", err)
	}

	app := &App{
		Config:  cfg,
		Manager: manager,
		Store:   store,
		Signer:  signer,
		RPC:     rpcClients,
	}

	return app, nil
}

// StartDepositListeners 启动充值监听器
func (a *App) StartDepositListeners(ctx context.Context) error {
	// TODO: 实现扫描器并启动监听
	// 这里需要根据配置创建扫描器并注册到 Manager
	log.Println("Deposit listeners started (placeholder)")
	return nil
}

// registerDefaultAssets 注册默认资产
func registerDefaultAssets(ctx context.Context, manager *service.Manager) error {
	assets := []domain.Asset{
		{
			Symbol:   "ETH",
			Chain:    domain.ChainEVM,
			Decimals: 18,
		},
		{
			Symbol:   "BTC",
			Chain:    domain.ChainBitcoin,
			Decimals: 8,
		},
	}

	for _, asset := range assets {
		if err := manager.RegisterAsset(ctx, asset); err != nil {
			// 如果已存在，忽略错误
			log.Printf("Asset %s already registered or error: %v", asset.Symbol, err)
		}
	}

	return nil
}

// MemoryKeyStore 简单的内存 KeyStore 实现
type MemoryKeyStore struct {
	keys map[string][]byte
}

// NewMemoryKeyStore 创建内存 KeyStore
func NewMemoryKeyStore() *MemoryKeyStore {
	return &MemoryKeyStore{
		keys: make(map[string][]byte),
	}
}

func (k *MemoryKeyStore) SaveKey(ctx context.Context, address string, encryptedKey []byte) error {
	if k.keys == nil {
		k.keys = make(map[string][]byte)
	}
	k.keys[address] = encryptedKey
	return nil
}

func (k *MemoryKeyStore) GetKey(ctx context.Context, address string) ([]byte, error) {
	if k.keys == nil {
		return nil, fmt.Errorf("key not found")
	}
	key, ok := k.keys[address]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}
	return key, nil
}

func (k *MemoryKeyStore) DeleteKey(ctx context.Context, address string) error {
	if k.keys == nil {
		return nil
	}
	delete(k.keys, address)
	return nil
}

