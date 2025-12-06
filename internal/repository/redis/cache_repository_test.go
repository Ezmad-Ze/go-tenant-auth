package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestCacheRepositoryBasicOperations tests basic cache operations
func TestCacheRepositoryBasicOperations(t *testing.T) {
	// Skip if Redis is not available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	repo := NewCacheRepository(client)

	t.Run("Set and Get", func(t *testing.T) {
		key := "test:key:1"
		value := map[string]string{"name": "test", "value": "123"}

		// Set value
		err := repo.Set(ctx, key, value, time.Minute)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Get value
		var result map[string]string
		err = repo.Get(ctx, key, &result)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if result["name"] != "test" || result["value"] != "123" {
			t.Errorf("Expected %v, got %v", value, result)
		}

		// Cleanup
		_ = repo.Delete(ctx, key)
	})

	t.Run("Exists", func(t *testing.T) {
		key := "test:key:2"

		// Should not exist initially
		exists, err := repo.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Key should not exist")
		}

		// Set value
		_ = repo.Set(ctx, key, "value", time.Minute)

		// Should exist now
		exists, err = repo.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Key should exist")
		}

		// Cleanup
		_ = repo.Delete(ctx, key)
	})

	t.Run("Increment", func(t *testing.T) {
		key := "test:counter:1"

		// Increment 3 times
		for i := 0; i < 3; i++ {
			count, err := repo.Increment(ctx, key, time.Minute)
			if err != nil {
				t.Fatalf("Increment failed: %v", err)
			}
			if count != int64(i+1) {
				t.Errorf("Expected count %d, got %d", i+1, count)
			}
		}

		// Cleanup
		_ = repo.Delete(ctx, key)
	})

	t.Run("Invalid TTL", func(t *testing.T) {
		key := "test:key:3"

		// Should reject zero TTL
		err := repo.Set(ctx, key, "value", 0)
		if err != ErrInvalidTTL {
			t.Errorf("Expected ErrInvalidTTL, got %v", err)
		}

		// Should reject negative TTL
		err = repo.Set(ctx, key, "value", -time.Second)
		if err != ErrInvalidTTL {
			t.Errorf("Expected ErrInvalidTTL, got %v", err)
		}
	})
}
