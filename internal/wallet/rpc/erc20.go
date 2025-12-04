package rpc

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TokenInfo 代币信息
type TokenInfo struct {
	Symbol   string
	Decimals uint8
	CachedAt time.Time
}

// ERC20Parser ERC20 事件解析器
type ERC20Parser struct {
	client     *ethclient.Client
	tokenCache map[string]*TokenInfo // 代币信息缓存
	cacheMu    sync.RWMutex          // 缓存读写锁
	cacheTTL   time.Duration         // 缓存过期时间
}

// NewERC20Parser 创建 ERC20 解析器
func NewERC20Parser(client *ethclient.Client) *ERC20Parser {
	return &ERC20Parser{
		client:     client,
		tokenCache: make(map[string]*TokenInfo),
		cacheTTL:   24 * time.Hour, // 默认缓存 24 小时
	}
}

// SetCacheTTL 设置缓存过期时间
func (p *ERC20Parser) SetCacheTTL(ttl time.Duration) {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.cacheTTL = ttl
}

// ParseTransferEvents 从交易回执中解析 Transfer 事件
func (p *ERC20Parser) ParseTransferEvents(ctx context.Context, txHash string) ([]*TokenTransfer, error) {
	hash := common.HexToHash(txHash)
	receipt, err := p.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	if receipt == nil || len(receipt.Logs) == 0 {
		return nil, nil
	}

	var transfers []*TokenTransfer
	var tokenAddresses []string
	tokenAddressMap := make(map[string]bool)

	// Transfer 事件的签名哈希
	// keccak256("Transfer(address,address,uint256)")
	transferEventSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	// 第一遍：收集所有代币地址
	for _, log := range receipt.Logs {
		if len(log.Topics) < 3 {
			continue
		}
		if log.Topics[0] != transferEventSig {
			continue
		}
		tokenAddress := log.Address.Hex()
		if !tokenAddressMap[tokenAddress] {
			tokenAddresses = append(tokenAddresses, tokenAddress)
			tokenAddressMap[tokenAddress] = true
		}
	}

	// 批量获取代币信息
	tokenInfoMap, _ := p.BatchGetTokenInfo(ctx, tokenAddresses)

	// 第二遍：解析所有 Transfer 事件
	for _, log := range receipt.Logs {
		// 检查是否是 Transfer 事件
		// topic[0] = keccak256("Transfer(address,address,uint256)")
		// topic[1] = from (indexed address)
		// topic[2] = to (indexed address)
		// data = value (uint256)
		if len(log.Topics) < 3 {
			continue
		}

		// 验证是否是 Transfer 事件
		if log.Topics[0] != transferEventSig {
			continue
		}

		// 解析事件数据
		from := common.BytesToAddress(log.Topics[1].Bytes())
		to := common.BytesToAddress(log.Topics[2].Bytes())

		// 解析 value（在 data 字段中，32字节）
		value := new(big.Int)
		if len(log.Data) >= 32 {
			value.SetBytes(log.Data[:32])
		} else {
			continue // 数据格式不正确
		}

		// 获取代币地址（合约地址）
		tokenAddress := log.Address.Hex()

		// 从批量获取的结果中获取代币信息
		var symbol string
		var decimals uint8
		if info, ok := tokenInfoMap[tokenAddress]; ok {
			symbol = info.Symbol
			decimals = info.Decimals
		} else {
			// 如果批量获取失败，回退到单个获取
			symbol, decimals, err = p.getTokenInfoWithCache(ctx, tokenAddress)
			if err != nil {
				symbol = "UNKNOWN"
				decimals = 18
			}
		}

		transfers = append(transfers, &TokenTransfer{
			TokenAddress: tokenAddress,
			From:         from.Hex(),
			To:           to.Hex(),
			Amount:       value,
			Symbol:       symbol,
			Decimals:     decimals,
		})
	}

	return transfers, nil
}

// getTokenInfoWithCache 获取代币信息（带缓存）
func (p *ERC20Parser) getTokenInfoWithCache(ctx context.Context, tokenAddress string) (string, uint8, error) {
	// 先检查缓存
	p.cacheMu.RLock()
	if cached, ok := p.tokenCache[tokenAddress]; ok {
		// 检查缓存是否过期
		if time.Since(cached.CachedAt) < p.cacheTTL {
			symbol := cached.Symbol
			decimals := cached.Decimals
			p.cacheMu.RUnlock()
			return symbol, decimals, nil
		}
	}
	p.cacheMu.RUnlock()

	// 缓存未命中或已过期，从链上获取
	symbol, decimals, err := p.getTokenInfo(ctx, tokenAddress)
	if err != nil {
		return "", 0, err
	}

	// 更新缓存
	p.cacheMu.Lock()
	p.tokenCache[tokenAddress] = &TokenInfo{
		Symbol:   symbol,
		Decimals: decimals,
		CachedAt: time.Now(),
	}
	p.cacheMu.Unlock()

	return symbol, decimals, nil
}

// getTokenInfo 获取代币信息（symbol, decimals），带重试机制
func (p *ERC20Parser) getTokenInfo(ctx context.Context, tokenAddress string) (string, uint8, error) {
	addr := common.HexToAddress(tokenAddress)

	// ERC20 标准函数签名
	// symbol() -> 0x95d89b41
	// decimals() -> 0x313ce567
	symbolData := common.Hex2Bytes("95d89b41")
	decimalsData := common.Hex2Bytes("313ce567")

	// 调用 symbol()，带重试
	var symbol string
	var symbolErr error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		symbolResult := ethereum.CallMsg{
			To:   &addr,
			Data: symbolData,
		}
		symbolBytes, err := p.client.CallContract(ctx, symbolResult, nil)
		if err == nil {
			symbol = parseStringFromBytes(symbolBytes)
			symbolErr = nil
			break
		}
		symbolErr = err
		if i < maxRetries-1 {
			// 等待后重试（指数退避）
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}
	if symbolErr != nil {
		return "", 0, fmt.Errorf("failed to call symbol after %d retries: %w", maxRetries, symbolErr)
	}

	// 调用 decimals()，带重试
	var decimals uint8 = 18 // 默认值
	var decimalsErr error
	for i := 0; i < maxRetries; i++ {
		decimalsResult := ethereum.CallMsg{
			To:   &addr,
			Data: decimalsData,
		}
		decimalsBytes, err := p.client.CallContract(ctx, decimalsResult, nil)
		if err == nil {
			if len(decimalsBytes) >= 32 {
				decimals = uint8(new(big.Int).SetBytes(decimalsBytes[31:32]).Uint64())
			}
			decimalsErr = nil
			break
		}
		decimalsErr = err
		if i < maxRetries-1 {
			// 等待后重试（指数退避）
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}
	if decimalsErr != nil {
		// decimals 失败时返回 symbol 和默认 decimals
		return symbol, 18, fmt.Errorf("failed to call decimals after %d retries: %w", maxRetries, decimalsErr)
	}

	return symbol, decimals, nil
}

// parseStringFromBytes 从 bytes32 解析字符串
// ERC20 symbol() 返回的是 string，ABI 编码格式：
// - 如果长度 <= 31: 直接在前32字节中，最后1字节是长度
// - 如果长度 > 31: 使用偏移量，前32字节是偏移量，接下来32字节是长度，然后是字符串内容
func parseStringFromBytes(data []byte) string {
	if len(data) < 32 {
		return ""
	}

	// 检查是否是短字符串格式（长度在最后1字节）
	length := int(data[31])
	if length > 0 && length <= 31 {
		// 短字符串，直接提取
		return strings.TrimRight(string(data[:length]), "\x00")
	}

	// 检查是否是长字符串格式（使用偏移量）
	if len(data) >= 64 {
		offset := new(big.Int).SetBytes(data[:32]).Uint64()
		if offset > 0 && offset < uint64(len(data)) {
			// 读取长度
			if offset+32 <= uint64(len(data)) {
				strLen := new(big.Int).SetBytes(data[offset : offset+32]).Uint64()
				if strLen > 0 && strLen < 256 && offset+32+strLen <= uint64(len(data)) {
					return string(data[offset+32 : offset+32+strLen])
				}
			}
		}
	}

	// 尝试直接解析（去除尾部的零字节）
	result := strings.TrimRight(string(data), "\x00")
	if len(result) > 0 {
		return result
	}

	return ""
}

// ClearCache 清空缓存
func (p *ERC20Parser) ClearCache() {
	p.cacheMu.Lock()
	defer p.cacheMu.Unlock()
	p.tokenCache = make(map[string]*TokenInfo)
}

// GetCacheSize 获取缓存大小
func (p *ERC20Parser) GetCacheSize() int {
	p.cacheMu.RLock()
	defer p.cacheMu.RUnlock()
	return len(p.tokenCache)
}

// BatchGetTokenInfo 批量获取代币信息（优化性能）
func (p *ERC20Parser) BatchGetTokenInfo(ctx context.Context, tokenAddresses []string) (map[string]*TokenInfo, error) {
	result := make(map[string]*TokenInfo)

	// 先检查缓存
	p.cacheMu.RLock()
	uncached := make([]string, 0)
	for _, addr := range tokenAddresses {
		if cached, ok := p.tokenCache[addr]; ok {
			if time.Since(cached.CachedAt) < p.cacheTTL {
				result[addr] = cached
			} else {
				uncached = append(uncached, addr)
			}
		} else {
			uncached = append(uncached, addr)
		}
	}
	p.cacheMu.RUnlock()

	// 批量获取未缓存的代币信息
	for _, addr := range uncached {
		symbol, decimals, err := p.getTokenInfo(ctx, addr)
		if err != nil {
			// 如果获取失败，使用默认值
			result[addr] = &TokenInfo{
				Symbol:   "UNKNOWN",
				Decimals: 18,
				CachedAt: time.Now(),
			}
			continue
		}

		info := &TokenInfo{
			Symbol:   symbol,
			Decimals: decimals,
			CachedAt: time.Now(),
		}
		result[addr] = info

		// 更新缓存
		p.cacheMu.Lock()
		p.tokenCache[addr] = info
		p.cacheMu.Unlock()
	}

	return result, nil
}
