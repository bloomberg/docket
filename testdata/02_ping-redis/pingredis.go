package pingredis

import (
	"github.com/go-redis/redis"
)

// PingRedis pings the Redis instance at redisAddr.
func PingRedis(redisAddr string) (string, error) {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})

	return client.Ping().Result()
}
