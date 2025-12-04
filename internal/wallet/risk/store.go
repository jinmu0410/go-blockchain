package risk

import (
	"context"
	"time"
)

// AddressListRepository 地址列表（白名单/黑名单）存储接口
type AddressListRepository interface {
	// AddToWhitelist 添加到白名单
	AddToWhitelist(ctx context.Context, address string, chain string, remark string) error
	// RemoveFromWhitelist 从白名单移除
	RemoveFromWhitelist(ctx context.Context, address string, chain string) error
	// IsWhitelisted 检查是否在白名单
	IsWhitelisted(ctx context.Context, address string, chain string) (bool, error)
	// ListWhitelist 列出白名单
	ListWhitelist(ctx context.Context, chain string, limit, offset int) ([]AddressListEntry, error)

	// AddToBlacklist 添加到黑名单
	AddToBlacklist(ctx context.Context, address string, chain string, remark string) error
	// RemoveFromBlacklist 从黑名单移除
	RemoveFromBlacklist(ctx context.Context, address string, chain string) error
	// IsBlacklisted 检查是否在黑名单
	IsBlacklisted(ctx context.Context, address string, chain string) (bool, error)
	// ListBlacklist 列出黑名单
	ListBlacklist(ctx context.Context, chain string, limit, offset int) ([]AddressListEntry, error)
}

// AddressListEntry 地址列表条目
type AddressListEntry struct {
	Address   string
	Chain     string
	Remark    string
	CreatedAt time.Time
	CreatedBy string
}

// RiskLogRepository 风控日志存储接口
type RiskLogRepository interface {
	// SaveLog 保存风控日志
	SaveLog(ctx context.Context, log *RiskLog) error
	// GetLog 获取单条日志
	GetLog(ctx context.Context, id string) (*RiskLog, error)
	// ListLogs 列出日志
	ListLogs(ctx context.Context, filter *RiskLogFilter) ([]*RiskLog, error)
}

// RiskLog 风控日志
type RiskLog struct {
	ID            string
	WithdrawalID  string
	UserID        string
	AssetSymbol   string
	ToAddress     string
	Amount        string
	Score         float64
	Approved      bool
	Remarks       string
	Rules         map[string]interface{} // 规则详情
	Decision      string                 // auto_approved, manual_review, rejected
	CreatedAt     time.Time
	Metadata      map[string]string
}

// RiskLogFilter 日志过滤条件
type RiskLogFilter struct {
	UserID      string
	WithdrawalID string
	Approved    *bool
	MinScore    *float64
	MaxScore    *float64
	StartTime   *time.Time
	EndTime     *time.Time
	Limit       int
	Offset      int
}

// InMemoryAddressListStore 内存实现的地址列表存储
type InMemoryAddressListStore struct {
	whitelist map[string]map[string]*AddressListEntry // chain -> address -> entry
	blacklist map[string]map[string]*AddressListEntry // chain -> address -> entry
}

// NewInMemoryAddressListStore 创建内存地址列表存储
func NewInMemoryAddressListStore() *InMemoryAddressListStore {
	return &InMemoryAddressListStore{
		whitelist: make(map[string]map[string]*AddressListEntry),
		blacklist: make(map[string]map[string]*AddressListEntry),
	}
}

func (s *InMemoryAddressListStore) AddToWhitelist(ctx context.Context, address string, chain string, remark string) error {
	if s.whitelist[chain] == nil {
		s.whitelist[chain] = make(map[string]*AddressListEntry)
	}
	s.whitelist[chain][address] = &AddressListEntry{
		Address:   address,
		Chain:     chain,
		Remark:    remark,
		CreatedAt: time.Now(),
	}
	return nil
}

func (s *InMemoryAddressListStore) RemoveFromWhitelist(ctx context.Context, address string, chain string) error {
	if s.whitelist[chain] != nil {
		delete(s.whitelist[chain], address)
	}
	return nil
}

func (s *InMemoryAddressListStore) IsWhitelisted(ctx context.Context, address string, chain string) (bool, error) {
	if s.whitelist[chain] != nil {
		_, ok := s.whitelist[chain][address]
		return ok, nil
	}
	return false, nil
}

func (s *InMemoryAddressListStore) ListWhitelist(ctx context.Context, chain string, limit, offset int) ([]AddressListEntry, error) {
	var result []AddressListEntry
	if s.whitelist[chain] != nil {
		count := 0
		for _, entry := range s.whitelist[chain] {
			if count < offset {
				count++
				continue
			}
			if len(result) >= limit {
				break
			}
			result = append(result, *entry)
		}
	}
	return result, nil
}

func (s *InMemoryAddressListStore) AddToBlacklist(ctx context.Context, address string, chain string, remark string) error {
	if s.blacklist[chain] == nil {
		s.blacklist[chain] = make(map[string]*AddressListEntry)
	}
	s.blacklist[chain][address] = &AddressListEntry{
		Address:   address,
		Chain:     chain,
		Remark:    remark,
		CreatedAt: time.Now(),
	}
	return nil
}

func (s *InMemoryAddressListStore) RemoveFromBlacklist(ctx context.Context, address string, chain string) error {
	if s.blacklist[chain] != nil {
		delete(s.blacklist[chain], address)
	}
	return nil
}

func (s *InMemoryAddressListStore) IsBlacklisted(ctx context.Context, address string, chain string) (bool, error) {
	if s.blacklist[chain] != nil {
		_, ok := s.blacklist[chain][address]
		return ok, nil
	}
	return false, nil
}

func (s *InMemoryAddressListStore) ListBlacklist(ctx context.Context, chain string, limit, offset int) ([]AddressListEntry, error) {
	var result []AddressListEntry
	if s.blacklist[chain] != nil {
		count := 0
		for _, entry := range s.blacklist[chain] {
			if count < offset {
				count++
				continue
			}
			if len(result) >= limit {
				break
			}
			result = append(result, *entry)
		}
	}
	return result, nil
}

// InMemoryRiskLogStore 内存实现的风控日志存储
type InMemoryRiskLogStore struct {
	logs map[string]*RiskLog
}

// NewInMemoryRiskLogStore 创建内存风控日志存储
func NewInMemoryRiskLogStore() *InMemoryRiskLogStore {
	return &InMemoryRiskLogStore{
		logs: make(map[string]*RiskLog),
	}
}

func (s *InMemoryRiskLogStore) SaveLog(ctx context.Context, log *RiskLog) error {
	if log.ID == "" {
		log.ID = log.WithdrawalID // 使用提现ID作为日志ID
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	s.logs[log.ID] = log
	return nil
}

func (s *InMemoryRiskLogStore) GetLog(ctx context.Context, id string) (*RiskLog, error) {
	log, ok := s.logs[id]
	if !ok {
		return nil, nil
	}
	return log, nil
}

func (s *InMemoryRiskLogStore) ListLogs(ctx context.Context, filter *RiskLogFilter) ([]*RiskLog, error) {
	var result []*RiskLog
	for _, log := range s.logs {
		if filter != nil {
			if filter.UserID != "" && log.UserID != filter.UserID {
				continue
			}
			if filter.WithdrawalID != "" && log.WithdrawalID != filter.WithdrawalID {
				continue
			}
			if filter.Approved != nil && log.Approved != *filter.Approved {
				continue
			}
			if filter.MinScore != nil && log.Score < *filter.MinScore {
				continue
			}
			if filter.MaxScore != nil && log.Score > *filter.MaxScore {
				continue
			}
			if filter.StartTime != nil && log.CreatedAt.Before(*filter.StartTime) {
				continue
			}
			if filter.EndTime != nil && log.CreatedAt.After(*filter.EndTime) {
				continue
			}
		}
		result = append(result, log)
	}
	
	// 限制返回数量
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(result) {
			result = result[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(result) {
			result = result[:filter.Limit]
		}
	}
	
	return result, nil
}

