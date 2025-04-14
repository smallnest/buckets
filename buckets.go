package buckets

import (
	"sort"
	"sync"
)

// Buckets 是一个泛型数据结构，用于存储基于时间戳的值
type Buckets[T Value] struct {
	buckets map[int64]*Bucket[T] // 使用map存储桶，键为时间戳
	mutex   sync.RWMutex         // 读写锁，保证线程安全
}

// NewBuckets 创建一个新的Buckets实例
func NewBuckets[T Value]() *Buckets[T] {
	return &Buckets[T]{
		buckets: make(map[int64]*Bucket[T]),
	}
}

// Add 将值添加到指定时间戳的桶中
func (b *Buckets[T]) Add(ts int64, value T) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	bucket, exists := b.buckets[ts]
	if !exists {
		bucket = NewBucket[T](ts)
		b.buckets[ts] = bucket
	}

	bucket.Add(value)
}

// Get 获取指定时间戳的桶
func (b *Buckets[T]) Get(ts int64) *Bucket[T] {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.buckets[ts] // 如果不存在将返回nil
}

// GetOldest 获取最早的时间戳桶
func (b *Buckets[T]) GetOldest() *Bucket[T] {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if len(b.buckets) == 0 {
		return nil
	}

	var oldestTS int64 = 1<<63 - 1 // 最大的int64值
	for ts := range b.buckets {
		if ts < oldestTS {
			oldestTS = ts
		}
	}

	return b.buckets[oldestTS]
}

// GetLatest 获取最新的时间戳桶
func (b *Buckets[T]) GetLatest() *Bucket[T] {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if len(b.buckets) == 0 {
		return nil
	}

	var latestTS int64 = -1 << 63 // 最小的int64值
	for ts := range b.buckets {
		if ts > latestTS {
			latestTS = ts
		}
	}

	return b.buckets[latestTS]
}

// GetAll 获取所有桶，按时间戳排序
func (b *Buckets[T]) GetAll() []*Bucket[T] {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	result := make([]*Bucket[T], 0, len(b.buckets))
	for _, bucket := range b.buckets {
		result = append(result, bucket)
	}

	// 按时间戳排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp < result[j].Timestamp
	})

	return result
}

// Remove 删除指定时间戳的桶
func (b *Buckets[T]) Remove(ts int64) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	_, exists := b.buckets[ts]
	if exists {
		delete(b.buckets, ts)
		return true
	}
	return false
}

// Count 返回桶的数量
func (b *Buckets[T]) Count() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return len(b.buckets)
}
