package internals

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}
type AuthService struct {
	config Config
	redis  *redis.Client
}

func NewAuthService(config Config, redisClient *redis.Client) (*AuthService, error) {
	if redisClient == nil {
		return nil, errors.New("redis client cannot be nil")
	}
	return &AuthService{
		config: config,
		redis:  redisClient,
	}, nil
}
func (s *AuthService) ApplyLoginDelay(ctx context.Context, username string) error {
	failedKey := fmt.Sprintf("failed_attempts:%s", username)
	failedAttempts, err := s.redis.Get(ctx, failedKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("internal server error: %w", err)
	}
	if failedAttempts > 0 {
		delay := s.calculateDelay(failedAttempts)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
func (s *AuthService) RecordFailedAttempt(ctx context.Context, username string) error {
	failedKey := fmt.Sprintf("failed_attempts:%s", username)
	failedAttempts, err := s.redis.Get(ctx, failedKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("internal server error: %v", err)
	}

	newAttempts := failedAttempts + 1
	if newAttempts >= s.config.MaxAttempts {
		return errors.New("account locked due to too many failed attempts")
	}
	if err := s.redis.Set(ctx, failedKey, newAttempts, time.Hour).Err(); err != nil {
		return fmt.Errorf("internal server error: %v", err)
	}
	return nil
}
func (s *AuthService) ResetFailedAttempts(ctx context.Context, username string) error {
	failedKey := fmt.Sprintf("failed_attempts:%s", username)
	if err := s.redis.Del(ctx, failedKey).Err(); err != nil {
		return fmt.Errorf("internal server error: %v", err)
	}
	return nil
}
func (s *AuthService) calculateDelay(failedAttempts int) time.Duration {
	delay := time.Duration(float64(s.config.InitialDelay) * math.Pow(2, float64(failedAttempts-1)))
	if delay > s.config.MaxDelay {
		return s.config.MaxDelay
	}
	return delay
}
