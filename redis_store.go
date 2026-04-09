package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore is a URLStore backed by Redis.
//
// Instead of holding data in a Go map (which disappears when the server
// restarts), we store everything in Redis. This means:
//   - Data survives server restarts
//   - Multiple instances of our app can share the same store
//   - We can set expiry on URLs (e.g. "expire after 30 days")
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore connects to Redis and returns a RedisStore.
// addr is in "host:port" format, e.g. "localhost:6379".
func NewRedisStore(addr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		// Password: "", // set this if your Redis requires authentication
		// DB:       0,  // Redis has 16 databases (0–15); 0 is the default
	})

	// Ping verifies the connection is alive. context.Background() is Go's
	// way of saying "no deadline, no cancellation" — just do it.
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("could not connect to Redis at %s: %w", addr, err)
	}

	return &RedisStore{client: client}, nil
}

// redisKey builds the Redis key for a short code.
// Namespacing keys with a prefix (e.g. "url:") is a Redis best practice —
// it prevents collisions if you ever store other data in the same Redis instance.
func redisKey(code string) string {
	return "url:" + code
}

// Save stores a new short code → long URL mapping as a Redis Hash.
//
// A Redis Hash is like a mini-dictionary inside a key:
//
//	redis> HSET url:aB3xZ9 long_url "https://hi-im-nina.github.io" created_at "..." clicks 0
//
// We use HSet (hash set) rather than Set (plain string) because we need to
// store multiple fields (long_url, created_at, clicks) under one key.
func (s *RedisStore) Save(code, longURL string) {
	ctx := context.Background()
	key := redisKey(code)

	s.client.HSet(ctx, key, map[string]any{
		"long_url":   longURL,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"clicks":     0,
	})
}

// Get retrieves a URLEntry from Redis by short code.
//
// HGetAll fetches every field of a hash at once. If the key doesn't exist,
// it returns an empty map — that's how we know the code wasn't found.
func (s *RedisStore) Get(code string) (*URLEntry, bool) {
	ctx := context.Background()
	key := redisKey(code)

	// HGetAll returns map[string]string of all fields in the hash.
	fields, err := s.client.HGetAll(ctx, key).Result()
	if err != nil || len(fields) == 0 {
		// len == 0 means the key doesn't exist in Redis
		return nil, false
	}

	createdAt, _ := time.Parse(time.RFC3339, fields["created_at"])
	clicks, _ := strconv.Atoi(fields["clicks"])

	return &URLEntry{
		LongURL:   fields["long_url"],
		CreatedAt: createdAt,
		Clicks:    clicks,
	}, true
}

// IncrementClicks atomically adds 1 to the clicks field.
//
// "Atomically" is the key word here — Redis processes commands one at a time,
// so HIncrBy is guaranteed to never have a race condition. This is one of
// Redis's killer features: you don't need mutexes like in our MemoryStore.
func (s *RedisStore) IncrementClicks(code string) {
	ctx := context.Background()
	s.client.HIncrBy(ctx, redisKey(code), "clicks", 1)
}

// Exists checks if a short code already has an entry in Redis.
// EXISTS returns 1 if the key exists, 0 if not.
func (s *RedisStore) Exists(code string) bool {
	ctx := context.Background()
	count, err := s.client.Exists(ctx, redisKey(code)).Result()
	return err == nil && count > 0
}
