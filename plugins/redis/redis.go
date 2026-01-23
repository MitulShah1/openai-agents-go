package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/redis/go-redis/v9"

	"github.com/MitulShah1/openai-agents-go/session"
)

// Store implements session.Session using Redis.
type Store struct {
	client     *redis.Client
	prefix     string
	expiration time.Duration
}

// Ensure Store implements session.Session
var _ session.Session = (*Store)(nil)

// Options configuration for Redis store
type Options struct {
	Addr       string
	Password   string
	DB         int
	Prefix     string        // Key prefix, default "agent:session:"
	Expiration time.Duration // TTL for sessions, default 24h
}

// New creates a new Redis session store.
func New(opts Options) (*Store, error) {
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

	return &Store{
		client:     client,
		prefix:     prefix,
		expiration: expiration,
	}, nil
}

// NewFromClient creates a new store using an existing redis client.
func NewFromClient(client *redis.Client, prefix string, expiration time.Duration) *Store {
	if prefix == "" {
		prefix = "agent:session:"
	}
	if expiration == 0 {
		expiration = 24 * time.Hour
	}
	return &Store{
		client:     client,
		prefix:     prefix,
		expiration: expiration,
	}
}

func (s *Store) key(sessionID string) string {
	return s.prefix + sessionID
}

// Get retrieves messages from Redis.
func (s *Store) Get(ctx context.Context, sessionID string) ([]openai.ChatCompletionMessageParamUnion, error) {
	key := s.key(sessionID)
	val, err := s.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, &session.StorageError{
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
func (s *Store) Append(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessageParamUnion) error {
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
		return &session.StorageError{
			SessionID: sessionID,
			Operation: "append",
			Err:       err,
		}
	}

	return nil
}

// Clear removes all messages from a session
func (s *Store) Clear(ctx context.Context, sessionID string) error {
	key := s.key(sessionID)
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return &session.StorageError{
			SessionID: sessionID,
			Operation: "clear",
			Err:       err,
		}
	}
	return nil
}

// Delete removes a session completely
func (s *Store) Delete(ctx context.Context, sessionID string) error {
	return s.Clear(ctx, sessionID)
}
