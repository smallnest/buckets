package buckets

import "sync"

type Value interface {
	Key() string
}

// Bucket 表示一个时间桶
type Bucket[T Value] struct {
	Timestamp int64

	mutex  sync.RWMutex
	Values map[string]Value
}

// NewBucket 创建一个新的时间桶
func NewBucket[T Value](timestamp int64) *Bucket[T] {
	return &Bucket[T]{
		Timestamp: timestamp,
		Values:    make(map[string]Value),
	}
}

// Add 向桶中添加一个值
func (b *Bucket[T]) Add(value T) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.Values[value.Key()] = value
}

// Add 向桶中添加一个值
func (b *Bucket[T]) AddOrUpdate(value T, updateFn func(T) T) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	v := b.Values[value.Key()]
	if v == nil {
		b.Values[value.Key()] = value
		return
	}

	v = updateFn(v.(T))
	b.Values[value.Key()] = v
}

// Get 从桶中获取一个值
func (b *Bucket[T]) Get(key string) (T, bool) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if value, ok := b.Values[key]; ok {
		return value.(T), true
	}
	var zero T
	return zero, false
}

// Remove 从桶中删除一个值
func (b *Bucket[T]) Remove(key string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	delete(b.Values, key)
}

// Update 更新桶中的值
func (b *Bucket[T]) Update(value T) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.Values[value.Key()]; ok {
		b.Values[value.Key()] = value
	}
}

// GetAll 获取桶中的所有值
func (b *Bucket[T]) GetAll() []T {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	values := make([]T, 0, len(b.Values))
	for _, value := range b.Values {
		values = append(values, value.(T))
	}
	return values
}
