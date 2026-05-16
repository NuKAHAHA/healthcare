package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const blacklistKeyPrefix = "bl:"

// TokenBlacklist stores revoked access-token JTIs in Redis.
// TTL equals the token's remaining lifetime so entries expire automatically.
type TokenBlacklist struct {
	rdb *redis.Client
}

func NewTokenBlacklist(rdb *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{rdb: rdb}
}

// Revoke stores a JTI in the blacklist until its natural expiry.
func (b *TokenBlacklist) Revoke(ctx context.Context, jti string, expAt time.Time) error {
	ttl := time.Until(expAt)
	if ttl <= 0 {
		// Already expired — nothing to blacklist
		return nil
	}
	key := blacklistKeyPrefix + jti
	if err := b.rdb.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("blacklist revoke: %w", err)
	}
	return nil
}

// IsRevoked returns true if the JTI is in the blacklist.
// On Redis error it returns (false, err) — callers should fail open.
func (b *TokenBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := b.rdb.Exists(ctx, blacklistKeyPrefix+jti).Result()
	if err != nil {
		return false, fmt.Errorf("blacklist check: %w", err)
	}
	return n > 0, nil
}
