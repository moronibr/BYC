package cache

import (
	"testing"
)

func TestCache_GetSetDelete(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test setting a value
	key := "testKey"
	value := []byte("testValue")
	if err := cache.Set(key, value); err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Test getting the value
	retrievedValue, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if string(retrievedValue) != string(value) {
		t.Errorf("Expected value %s, got %s", string(value), string(retrievedValue))
	}

	// Test deleting the value
	if err := cache.Delete(key); err != nil {
		t.Fatalf("Failed to delete value: %v", err)
	}

	// Test getting the deleted value
	_, err = cache.Get(key)
	if err == nil {
		t.Error("Expected error for deleted key")
	}
}

func TestCache_Reset(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set some values
	key1 := "key1"
	value1 := []byte("value1")
	key2 := "key2"
	value2 := []byte("value2")

	if err := cache.Set(key1, value1); err != nil {
		t.Fatalf("Failed to set value1: %v", err)
	}
	if err := cache.Set(key2, value2); err != nil {
		t.Fatalf("Failed to set value2: %v", err)
	}

	// Reset the cache
	if err := cache.Reset(); err != nil {
		t.Fatalf("Failed to reset cache: %v", err)
	}

	// Test getting the reset values
	_, err = cache.Get(key1)
	if err == nil {
		t.Error("Expected error for reset key1")
	}
	_, err = cache.Get(key2)
	if err == nil {
		t.Error("Expected error for reset key2")
	}
}

func TestCache_LenCapacity(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set some values
	key1 := "key1"
	value1 := []byte("value1")
	key2 := "key2"
	value2 := []byte("value2")

	if err := cache.Set(key1, value1); err != nil {
		t.Fatalf("Failed to set value1: %v", err)
	}
	if err := cache.Set(key2, value2); err != nil {
		t.Fatalf("Failed to set value2: %v", err)
	}

	// Test Len
	if cache.Len() != 2 {
		t.Errorf("Expected length 2, got %d", cache.Len())
	}

	// Test Capacity
	if cache.Capacity() <= 0 {
		t.Error("Expected positive capacity")
	}
}

func TestCache_Stats(t *testing.T) {
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Set some values
	key1 := "key1"
	value1 := []byte("value1")
	key2 := "key2"
	value2 := []byte("value2")

	if err := cache.Set(key1, value1); err != nil {
		t.Fatalf("Failed to set value1: %v", err)
	}
	if err := cache.Set(key2, value2); err != nil {
		t.Fatalf("Failed to set value2: %v", err)
	}

	// Test Stats
	stats := cache.Stats()
	if stats == nil {
		t.Error("Expected non-nil stats")
	}
}
