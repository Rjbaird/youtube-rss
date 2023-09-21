package db

import (
	"log"

	"github.com/bairrya/youtube-rss/config"
	"github.com/go-redis/redis"
	"github.com/nitishm/go-rejson"
)

func RedisConnect() (*rejson.Handler, error) {
	config, err := config.ENV()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	rson := rejson.NewReJSONHandler()

	opt, err := redis.ParseURL(config.REDIS_URL)
	if err != nil {
		log.Printf("Error parsing redis url: %s", config.REDIS_URL)
		return nil, err
	}
	
	client := redis.NewClient(opt)
	// defer func() {
	// 	if err := client.Close(); err != nil {
	// 		log.Fatalf("goredis - failed to communicate to redis-server: %v", err)
	// 	}
	// }()
	rson.SetGoRedisClient(client)
	return rson, nil
}
