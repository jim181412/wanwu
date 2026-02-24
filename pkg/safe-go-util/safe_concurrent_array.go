package safe_go_util

import (
	"sync"
)

// SafeArray 泛型并发安全数组
type SafeArray[T any] struct {
	mu    sync.RWMutex
	items []T
}

// NewSafeArray 创建泛型并发安全数组
func NewSafeArray[T any]() *SafeArray[T] {
	return &SafeArray[T]{
		items: make([]T, 0),
	}
}

// Append 添加元素
func (sa *SafeArray[T]) Append(item T) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.items = append(sa.items, item)
}

// Get 获取元素
func (sa *SafeArray[T]) Get(index int) (T, bool) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	var zero T
	if index < 0 || index >= len(sa.items) {
		return zero, false
	}
	return sa.items[index], true
}

// Length 获取元素数量
func (sa *SafeArray[T]) Length() int {
	return len(sa.items)
}

// Set 设置元素
func (sa *SafeArray[T]) Set(index int, value T) bool {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if index < 0 || index >= len(sa.items) {
		return false
	}
	sa.items[index] = value
	return true
}

// Find 查找元素（使用自定义比较函数）
func (sa *SafeArray[T]) Find(compareFunc func(T) bool) (T, bool) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	var zero T
	for _, item := range sa.items {
		if compareFunc(item) {
			return item, true
		}
	}
	return zero, false
}

// Range 便利元素（使用自定义比较函数）
func (sa *SafeArray[T]) Range(compareFunc func(T)) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	for _, item := range sa.items {
		compareFunc(item)
	}
}

// All 查找元素（使用自定义比较函数）所有都满足条件
func (sa *SafeArray[T]) All(compareFunc func(T) bool) bool {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	for _, item := range sa.items {
		if !compareFunc(item) {
			return false
		}
	}
	return true
}

// Filter 过滤元素
func (sa *SafeArray[T]) Filter(filterFunc func(T) bool) []T {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	result := make([]T, 0)
	for _, item := range sa.items {
		if filterFunc(item) {
			result = append(result, item)
		}
	}
	return result
}
