package internals

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Manager interface {
	NewTokenManager(redisClient *redis.Client, context context.Context, tokenTTL time.Duration) Manager
	IsTokenRevoked(token string) (bool, string, error)
	RevokeToken(token, sessionID string) error
	RevokeAccessToken(tokenString string, expirationTime time.Time) error
	IsAccessTokenRevoked(tokenString string) (bool, error)
}
type TokenManager struct {
	redisClient *redis.Client
	context     context.Context
	tokenTTL    time.Duration
}

func (tm *TokenManager) NewTokenManager(redisClient *redis.Client, context context.Context, tokenTTL time.Duration) Manager {
	return &TokenManager{
		redisClient: redisClient,
		context:     context,
		tokenTTL:    tokenTTL,
	}
}
func (tm *TokenManager) IsTokenRevoked(token string) (bool, string, error) {
	sessionId, err := tm.redisClient.Get(tm.context, token).Result()
	if err != nil {
		if err == redis.Nil {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check token: %w", err)
	}
	return true, sessionId, nil
}
func (tm *TokenManager) RevokeToken(token, sessionID string) error {
	err := tm.redisClient.Set(tm.context, token, sessionID, tm.tokenTTL)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %v", err)
	}
	return nil
}
func (tm *TokenManager) SetTokenExpiration(token string, duration time.Duration) error {
	return tm.redisClient.Expire(tm.context, token, duration).Err()
}

func (tm *TokenManager) GetTokenTTL(token string) (time.Duration, error) {
	return tm.redisClient.TTL(tm.context, token).Result()
}
func (tm *TokenManager) RevokeMultipleTokens(tokenMap map[string]string) error {
	pipe := tm.redisClient.Pipeline()
	for token, sessionID := range tokenMap {
		pipe.Set(tm.context, token, sessionID, tm.tokenTTL)
	}
	_, err := pipe.Exec(tm.context)
	return err
}
func (tm *TokenManager) RevokeAccessToken(tokenString string, expirationTime time.Time) error {
	ttl := time.Until(expirationTime)
	if ttl <= 0 {
		return nil
	}
	key := "access:" + tokenString
	return tm.redisClient.Set(tm.context, key, "revoked", ttl).Err()
}
func (tm *TokenManager) IsAccessTokenRevoked(tokenString string) (bool, error) {
	key := "access:" + tokenString
	_, err := tm.redisClient.Get(tm.context, key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
