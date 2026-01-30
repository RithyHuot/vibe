package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Set a value
	cache.Set("key1", "value1")

	// Get the value
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)
}

func TestCache_GetNonExistent(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Try to get a non-existent key
	value, found := cache.Get("nonexistent")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	// Set a value
	cache.Set("key1", "value1")

	// Immediately get it - should be found
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Try to get it again - should be expired
	value, found = cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Set a value
	cache.Set("key1", "value1")

	// Verify it exists
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Delete it
	cache.Delete("key1")

	// Verify it's gone
	value, found = cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Set multiple values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Clear all
	cache.Clear()

	// Verify all are gone
	_, found := cache.Get("key1")
	assert.False(t, found)
	_, found = cache.Get("key2")
	assert.False(t, found)
	_, found = cache.Get("key3")
	assert.False(t, found)
}

func TestCache_CleanExpired(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	// Set values at different times
	cache.Set("key1", "value1")
	time.Sleep(50 * time.Millisecond)
	cache.Set("key2", "value2")
	time.Sleep(60 * time.Millisecond)

	// key1 should be expired, key2 should still be valid
	cache.CleanExpired()

	_, found := cache.Get("key1")
	assert.False(t, found, "key1 should be expired and cleaned")

	value, found := cache.Get("key2")
	assert.True(t, found, "key2 should still be valid")
	assert.Equal(t, "value2", value)
}

func TestCache_UpdateValue(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Set initial value
	cache.Set("key1", "value1")

	// Update value
	cache.Set("key1", "value2")

	// Get updated value
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value2", value)
}

func TestCache_DifferentTypes(t *testing.T) {
	cache := NewCache(1 * time.Minute)

	// Store different types
	cache.Set("string", "hello")
	cache.Set("int", 42)
	cache.Set("bool", true)
	cache.Set("struct", struct{ Name string }{"test"})

	// Retrieve and verify
	value, found := cache.Get("string")
	assert.True(t, found)
	assert.Equal(t, "hello", value)

	value, found = cache.Get("int")
	assert.True(t, found)
	assert.Equal(t, 42, value)

	value, found = cache.Get("bool")
	assert.True(t, found)
	assert.Equal(t, true, value)

	value, found = cache.Get("struct")
	assert.True(t, found)
	assert.Equal(t, struct{ Name string }{"test"}, value)
}
