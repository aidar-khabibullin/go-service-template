package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

const ttl = time.Hour

type Redis struct {
	client *redis.Client
}

func New(client *redis.Client) *Redis {
	return &Redis{client: client}
}
