package buckets

import (
	"fmt"
	"sync"
	"testing"
)

// Mock Value implementation for testing
type TestValue struct {
	ID    int
	Value string
}

func (v TestValue) Key() string {
	return fmt.Sprintf("%d", v.ID)
}

// Helper function to create a buckets instance with test data
func createTestBuckets() *Buckets[TestValue] {
	b := NewBuckets[TestValue]()
	b.Add(100, TestValue{ID: 1, Value: "test1"})
	b.Add(200, TestValue{ID: 2, Value: "test2"})
	b.Add(300, TestValue{ID: 3, Value: "test3"})
	return b
}

// TestNewBuckets tests the creation of a new Buckets instance
func TestNewBuckets(t *testing.T) {
	b := NewBuckets[TestValue]()
	if b == nil {
		t.Fatal("NewBuckets returned nil")
	}
	if b.buckets == nil {
		t.Fatal("buckets map is nil")
	}
	if len(b.buckets) != 0 {
		t.Fatalf("expected empty buckets, got %d", len(b.buckets))
	}
}

// TestAdd tests the Add method
func TestAdd(t *testing.T) {
	b := NewBuckets[TestValue]()

	// Test adding to a new bucket
	ts := int64(100)
	value := TestValue{ID: 1, Value: "test1"}
	b.Add(ts, value)

	bucket := b.Get(ts)
	if bucket == nil {
		t.Fatalf("bucket not created for timestamp %d", ts)
	}
	if bucket.Timestamp != ts {
		t.Fatalf("expected timestamp %d, got %d", ts, bucket.Timestamp)
	}
	if len(bucket.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(bucket.Values))
	}

	// Test adding to an existing bucket
	value2 := TestValue{ID: 2, Value: "test2"}
	b.Add(ts, value2)

	bucket = b.Get(ts)
	if len(bucket.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(bucket.Values))
	}

	// Test concurrent adds
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Add(-1)
			b.Add(ts, TestValue{ID: i + 10, Value: "concurrent"})
		}(i)
	}
	wg.Wait()

	bucket = b.Get(ts)
	if len(bucket.Values) != 102 { // 2 previous + 100 concurrent
		t.Fatalf("expected 102 values after concurrent adds, got %d", len(bucket.Values))
	}
}

// TestGet tests the Get method
func TestGet(t *testing.T) {
	b := createTestBuckets()

	// Test getting an existing bucket
	bucket := b.Get(200)
	if bucket == nil {
		t.Fatal("failed to get existing bucket")
	}
	if bucket.Timestamp != 200 {
		t.Fatalf("expected timestamp 200, got %d", bucket.Timestamp)
	}

	// Test getting a non-existent bucket
	bucket = b.Get(999)
	if bucket != nil {
		t.Fatal("expected nil for non-existent bucket")
	}
}

// TestGetOldest tests the GetOldest method
func TestGetOldest(t *testing.T) {
	// Test with empty buckets
	b := NewBuckets[TestValue]()
	bucket := b.GetOldest()
	if bucket != nil {
		t.Fatal("expected nil for empty buckets")
	}

	// Test with populated buckets
	b = createTestBuckets()
	bucket = b.GetOldest()
	if bucket == nil {
		t.Fatal("GetOldest returned nil")
	}
	if bucket.Timestamp != 100 {
		t.Fatalf("expected oldest timestamp 100, got %d", bucket.Timestamp)
	}

	// Add older bucket and verify
	b.Add(50, TestValue{ID: 4, Value: "test4"})
	bucket = b.GetOldest()
	if bucket.Timestamp != 50 {
		t.Fatalf("expected oldest timestamp 50, got %d", bucket.Timestamp)
	}
}

// TestGetLatest tests the GetLatest method
func TestGetLatest(t *testing.T) {
	// Test with empty buckets
	b := NewBuckets[TestValue]()
	bucket := b.GetLatest()
	if bucket != nil {
		t.Fatal("expected nil for empty buckets")
	}

	// Test with populated buckets
	b = createTestBuckets()
	bucket = b.GetLatest()
	if bucket == nil {
		t.Fatal("GetLatest returned nil")
	}
	if bucket.Timestamp != 300 {
		t.Fatalf("expected latest timestamp 300, got %d", bucket.Timestamp)
	}

	// Add newer bucket and verify
	b.Add(400, TestValue{ID: 4, Value: "test4"})
	bucket = b.GetLatest()
	if bucket.Timestamp != 400 {
		t.Fatalf("expected latest timestamp 400, got %d", bucket.Timestamp)
	}
}

//
