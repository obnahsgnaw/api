package jwt

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/obnahsgnaw/application/pkg/utils"
	"strings"
	"time"
)

// GetUserJwtKey get user jwt key
func GetUserJwtKey(rds *redis.Client, subject string, id string) (string, error) {
	ck := tokenCacheKey(subject, id)
	rs := rds.Get(context.Background(), ck)
	if rs.Err() != nil {
		if rs.Err() == redis.Nil {
			return "", nil
		}
		return "", utils.NewWrappedError("get user jwt key failed", rs.Err())
	}
	return rs.Val(), nil
}

// SetUserJwtKey set user jwt key
func SetUserJwtKey(rds *redis.Client, subject, id, key string, ttl time.Duration) error {
	rs := rds.Set(context.Background(), tokenCacheKey(subject, id), key, ttl)
	if rs.Err() != nil {
		return utils.NewWrappedError("set user jwt key failed", rs.Err())
	}
	return nil
}

// ExpireUserJwtKey lengthen user jwt key
func ExpireUserJwtKey(rds *redis.Client, subject, id string, ttl time.Duration) error {
	rs := rds.Expire(context.Background(), tokenCacheKey(subject, id), ttl)
	if rs.Err() != nil {
		return utils.NewWrappedError("expire user jwt key failed", rs.Err())
	}
	return nil
}

// DelUserJwtKey delete user jwt key
func DelUserJwtKey(rds *redis.Client, subject, id string) error {
	rs := rds.Del(context.Background(), tokenCacheKey(subject, id))
	if rs.Err() != nil {
		return utils.NewWrappedError("del user jwt key failed", rs.Err())
	}
	return rs.Err()
}

func tokenCacheKey(subject, id string) string {
	return strings.Join([]string{`tokens`, subject, id}, ":")
}
