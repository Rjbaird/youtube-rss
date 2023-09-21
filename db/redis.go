package db

import (
	"log"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/bairrya/youtube-rss/config"
	"github.com/go-redis/redis"
	"github.com/nitishm/go-rejson"
)

func RedisStackConnect() (*redis.Client, *rejson.Handler, *redisearch.Client, error) {
	config, err := config.ENV()
	if err != nil {
		log.Fatal(err)
		return nil, nil, nil, err
	}
	opt, err := redis.ParseURL(config.REDIS_URL)
	if err != nil {
		log.Printf("Error parsing redis url: %s", config.REDIS_URL)
		return nil, nil, nil, err
	}
	
	
	client := redis.NewClient(opt)
	rson := rejson.NewReJSONHandler()
	search := redisearch.NewClient(config.REDIS_URL, "rss")
	// defer func() {
	// 	if err := client.Close(); err != nil {
	// 		log.Fatalf("goredis - failed to communicate to redis-server: %v", err)
	// 	}
	// }()
	rson.SetGoRedisClient(client)
	return client, rson, search, nil
}
