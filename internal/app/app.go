package app

import (
	"context"
	"fmt"
	"log"

	"github.com/jinmu/go-blockchain/internal/config"
	"github.com/jinmu/go-blockchain/internal/wallet/deposit"
	"github.com/jinmu/go-blockchain/internal/wallet/notification"
	"github.com/jinmu/go-blockchain/internal/wallet/rpc"
	"github.com/jinmu/go-blockchain/internal/wallet/risk"
	"github.com/jinmu/go-blockchain/internal/wallet/scanner"
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
	var riskController risk.Controller
	if cfg.Database.Type == "memory" {
		// 使用默认风控（开发测试）
		riskController = &risk.NoopController{}
	} else {
		// 使用完整风控系统
		riskController = risk.NewRiskController(store, nil)
	}

	// 5. 初始化通知管理器（可选）
	var notifierManager *notification.Manager
	if cfg.Database.Type != "memory" {
		// 生产环境启用通知
		webhookNotifier := notification.NewWebhookNotifier()
		configStore := notification.NewInMemoryConfigStore() // 可以替换为数据库存储
		notifierManager = notification.NewManager(webhookNotifier, configStore)
	}

	// 6. 初始化 Manager
	managerOpts := []service.Option{
		service.WithRiskController(riskController),
		service.WithSigner(signer),
	}
	if notifierManager != nil {
		managerOpts = append(managerOpts, service.WithNotifier(notifierManager))
	}
	manager := service.NewManager(store, managerOpts...)

	// 7. 注册默认资产
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
	// 为每个链创建扫描器
	for chainType, rpcClient := range a.RPC {
		var sc scanner.Scanner
		var confirmations, reorgDepth uint64

		// 根据链类型创建对应的扫描器
		switch chainType {
		case domain.ChainEVM:
			confirmations = 12
			reorgDepth = 12
			sc = scanner.NewEVMScanner(chainType, rpcClient, a.Store, confirmations, reorgDepth)
		case domain.ChainBitcoin:
			confirmations = 6
			reorgDepth = 6
			sc = scanner.NewBitcoinScanner(chainType, rpcClient, a.Store, confirmations, reorgDepth)
		case domain.ChainSolana:
			confirmations = 32
			reorgDepth = 32
			sc = scanner.NewSolanaScanner(chainType, rpcClient, a.Store, confirmations, reorgDepth)
		default:
			log.Printf("Unsupported chain type: %s, skipping scanner", chainType)
			continue
		}

		// 注册扫描器
		a.Manager.RegisterScanner(sc)

		// 启动扫描
		depositHandler := func(ctx context.Context, event domain.DepositEvent) error {
			return a.Manager.HandleDepositEvent(ctx, event)
		}

		reorgHandler := func(ctx context.Context, chain domain.ChainType, fromHeight, toHeight uint64) error {
			// 使用 Manager 的 RollbackDeposit 方法处理重组
			reorgHandler := deposit.NewDefaultReorgHandlerWithConfig(
				a.Store,
				a.Manager,
				nil, // Ledger（可选）
				deposit.DefaultReorgDepths,
			)
			return reorgHandler.OnReorg(ctx, chain, fromHeight, toHeight)
		}

		if err := sc.Subscribe(ctx, depositHandler, reorgHandler); err != nil {
			return fmt.Errorf("failed to start scanner for %s: %w", chainType, err)
		}

		log.Printf("Deposit listener started for chain: %s (confirmations: %d, reorg_depth: %d)", chainType, confirmations, reorgDepth)
	}

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

