package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

type Store struct {
	client *redis.Client
}

func NewStore(url string) (*Store, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Store{client: client}, nil
}

func (s *Store) Client() *redis.Client {
	return s.client
}

func sessionKey(customerNumber string) string {
	return fmt.Sprintf("session:%s:messages", customerNumber)
}

func summaryKey(customerNumber string) string {
	return fmt.Sprintf("session:%s:summary", customerNumber)
}

func (s *Store) SaveSessionMessages(ctx context.Context, customerNumber string, messages []llm.ChatMessage, ttl time.Duration) error {
	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}
	return s.client.Set(ctx, sessionKey(customerNumber), data, ttl).Err()
}

func (s *Store) GetSessionMessages(ctx context.Context, customerNumber string) ([]llm.ChatMessage, error) {
	data, err := s.client.Get(ctx, sessionKey(customerNumber)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var messages []llm.ChatMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
	}
	return messages, nil
}

func (s *Store) DeleteSession(ctx context.Context, customerNumber string) error {
	pipe := s.client.Pipeline()
	pipe.Del(ctx, sessionKey(customerNumber))
	pipe.Del(ctx, summaryKey(customerNumber))
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) SetSessionSummary(ctx context.Context, customerNumber string, summary string, ttl time.Duration) error {
	return s.client.Set(ctx, summaryKey(customerNumber), summary, ttl).Err()
}

func (s *Store) GetSessionSummary(ctx context.Context, customerNumber string) (string, error) {
	result, err := s.client.Get(ctx, summaryKey(customerNumber)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return result, nil
}

func (s *Store) RefreshSessionTTL(ctx context.Context, customerNumber string, ttl time.Duration) error {
	pipe := s.client.Pipeline()
	pipe.Expire(ctx, sessionKey(customerNumber), ttl)
	pipe.Expire(ctx, summaryKey(customerNumber), ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func rateLimitKey(customerNumber string, window string) string {
	return fmt.Sprintf("rl:%s:%s", customerNumber, window)
}

func muteKey(customerNumber string) string {
	return fmt.Sprintf("mute:%s", customerNumber)
}

func spamKey(customerNumber string) string {
	return fmt.Sprintf("spam:%s", customerNumber)
}

func (s *Store) IncrementRateLimit(ctx context.Context, customerNumber string, window string, ttl time.Duration) (int64, error) {
	key := rateLimitKey(customerNumber, window)
	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func (s *Store) GetRateLimitCount(ctx context.Context, customerNumber string, window string) (int64, error) {
	count, err := s.client.Get(ctx, rateLimitKey(customerNumber, window)).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func (s *Store) MuteNumber(ctx context.Context, customerNumber string, duration time.Duration) error {
	return s.client.Set(ctx, muteKey(customerNumber), "1", duration).Err()
}

func (s *Store) IsMuted(ctx context.Context, customerNumber string) (bool, error) {
	exists, err := s.client.Exists(ctx, muteKey(customerNumber)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *Store) UnmuteNumber(ctx context.Context, customerNumber string) error {
	return s.client.Del(ctx, muteKey(customerNumber)).Err()
}

func (s *Store) RecordMessageForSpamDetection(ctx context.Context, customerNumber string, messageHash string) (int64, error) {
	key := spamKey(customerNumber)
	pipe := s.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: messageHash,
	})
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", time.Now().Add(-60*time.Second).Unix()))
	count := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, 2*time.Minute)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return count.Val(), nil
}
