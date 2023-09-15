package db

import (
	"log"

	"github.com/bairrya/youtube-rss/config"
	"github.com/go-redis/redis"
)

func RedisConnect() (*redis.Client, error) {
	config, err := config.ENV()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	opt, err := redis.ParseURL(config.REDIS_URL)
	if err != nil {
		log.Printf("Error parsing redis url: %s", config.REDIS_URL)
		return nil, err
	}

	client := redis.NewClient(opt)
	return client, nil
}
