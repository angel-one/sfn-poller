package pollable

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	GetToken(ctx context.Context, arn string) (string, error)
	ReturnToken(ctx context.Context, token string) error
}

type RateLimiterImpl struct {
	redisClient     redis.UniversalClient
	rateLimitConfig map[string]int32
}

const getTokenRedisScript = `
local key = KEYS[1] + os.date(" %Y-%m-%d %H:%M:%S")
local max = tonumber(ARGV[1])
local count = redis.call("incr", key)
if count > max then
	redis.call("incrby", key, -1)
end
redis.call("expire", key, 10)
return key
`

const returnTokenRedisScript = `
local key = KEYS[1]
local count = redis.call("get", key)
if count < 0 then
	redis.call("set", key, 0)
end
redis.call("expire", key, 10)
return count
`

func NewRateLimiter(redisClient redis.UniversalClient, rateLimitConfig map[string]int32) RateLimiter {
	return &RateLimiterImpl{redisClient: redisClient, rateLimitConfig: rateLimitConfig}
}

func (rl *RateLimiterImpl) GetToken(ctx context.Context, arn string) (string, error) {
	maxLimit := rl.rateLimitConfig[arn]
	return redis.NewScript(getTokenRedisScript).Run(ctx, rl.redisClient, []string{arn}, maxLimit).Text()
}

func (rl *RateLimiterImpl) ReturnToken(ctx context.Context, token string) error {
	return redis.NewScript(returnTokenRedisScript).Run(ctx, rl.redisClient, []string{token}).Err()
}
