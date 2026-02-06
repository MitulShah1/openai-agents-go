//go:build redis

package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/redis/go-redis/v9"
)

// RedisStore implements Session using Redis.
type RedisStore struct {
	client     *redis.Client
	prefix     string
	expiration time.Duration
}

// Ensure RedisStore implements Session
var _ Session = (*RedisStore)(nil)

// RedisOptions configuration for Redis store
type RedisOptions struct {
	Addr       string
	Password   string
	DB         int
	Prefix     string        // Key prefix, default "agent:session:"
	Expiration time.Duration // TTL for sessions, default 24h
}

// NewRedisSession creates a new Redis session store.
func NewRedisSession(opts RedisOptions) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	prefix := opts.Prefix
	if prefix == "" {
		prefix = "agent:session:"
	}

	expiration := opts.Expiration
	if expiration == 0 {
		expiration = 24 * time.Hour
	}

	return &RedisStore{
		client:     client,
		prefix:     prefix,
		expiration: expiration,
	}, nil
}

// NewRedisSessionFromClient creates a new store using an existing redis client.
func NewRedisSessionFromClient(client *redis.Client, prefix string, expiration time.Duration) *RedisStore {
	if prefix == "" {
		prefix = "agent:session:"
	}
	if expiration == 0 {
		expiration = 24 * time.Hour
	}
	return &RedisStore{
		client:     client,
		prefix:     prefix,
		expiration: expiration,
	}
}

func (s *RedisStore) key(sessionID string) string {
	return s.prefix + sessionID
}

// Get retrieves messages from Redis.
func (s *RedisStore) Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	key := s.key(sessionID)
	val, err := s.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, &StorageError{
			SessionID: sessionID,
			Operation: "get",
			Err:       err,
		}
	}

	if len(val) == 0 {
		// No messages implies session might not exist or empty
		return nil, nil // Return empty, not error
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(val))
	for _, data := range val {
		var msg openai.ChatCompletionMessageParamUnion
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// Append adds messages to the session history.
func (s *RedisStore) Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
	if len(messages) == 0 {
		return nil
	}
	key := s.key(sessionID)

	pipeline := s.client.Pipeline()

	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		pipeline.RPush(ctx, key, data)
	}

	// Refresh expiration
	pipeline.Expire(ctx, key, s.expiration)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return &StorageError{
			SessionID: sessionID,
			Operation: "append",
			Err:       err,
		}
	}

	return nil
}

// Clear removes all messages from a session
func (s *RedisStore) Clear(ctx context.Context, sessionID string) error {
	key := s.key(sessionID)
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return &StorageError{
			SessionID: sessionID,
			Operation: "clear",
			Err:       err,
		}
	}
	return nil
}

// Delete removes a session completely
func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	return s.Clear(ctx, sessionID)
}

// Compact implements OpenAIResponsesCompactionAwareSession.
// It trims the session history to ensure it fits within limits.
// Currently implements a basic sliding window approach.
func (s *RedisStore) Compact(ctx context.Context, sessionID string, maxTokens int) error {
	// 1. Get all messages
	msgs, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return nil
	}

	// 2. Compact messages using DefaultCompactor (placeholder for now)
	compactedMsgs := DefaultCompactor(msgs, maxTokens)

	// If no change, return
	if len(compactedMsgs) == len(msgs) {
		return nil
	}

	// 3. Update session in Redis (Clear + Append)
	// We use a transaction (pipeline) to ensure atomicity if possible, but Clear+Append is simplest.
	// WATCH key? For now, simple overwrite.
	if err := s.Clear(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to clear session for compaction: %w", err)
	}

	if err := s.Append(ctx, sessionID, compactedMsgs); err != nil {
		return fmt.Errorf("failed to append compacted messages: %w", err)
	}

	return nil
}
