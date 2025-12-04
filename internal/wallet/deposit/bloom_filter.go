package deposit

import (
	"github.com/bits-and-blooms/bloom/v3"
)

// SimpleBloomFilter 简单的布隆过滤器实现
type SimpleBloomFilter struct {
	filter *bloom.BloomFilter
}

// NewSimpleBloomFilter 创建布隆过滤器
// capacity: 预期元素数量
// falsePositiveRate: 误报率（0-1）
func NewSimpleBloomFilter(capacity uint, falsePositiveRate float64) *SimpleBloomFilter {
	return &SimpleBloomFilter{
		filter: bloom.NewWithEstimates(capacity, falsePositiveRate),
	}
}

// Add 添加元素
func (b *SimpleBloomFilter) Add(item []byte) {
	if b.filter != nil {
		b.filter.Add(item)
	}
}

// Test 测试元素是否存在
func (b *SimpleBloomFilter) Test(item []byte) bool {
	if b.filter == nil {
		return true // 如果没有过滤器，默认返回 true（处理所有交易）
	}
	return b.filter.Test(item)
}

// NoopBloomFilter 空实现（用于开发测试，不过滤任何交易）
type NoopBloomFilter struct{}

func (n *NoopBloomFilter) Add(item []byte) {}
func (n *NoopBloomFilter) Test(item []byte) bool {
	return true // 总是返回 true，处理所有交易
}

