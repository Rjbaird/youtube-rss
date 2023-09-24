package db

import (
	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/go-redis/redis"
	"github.com/nitishm/go-rejson"
)

type RedisStack struct {
	Client *redis.Client
	ReJSON *rejson.Handler
	Search *redisearch.Client
}

type Channel struct {
	Handle      string   `json:"handle"`
	ChannelID   string   `json:"channel_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}
