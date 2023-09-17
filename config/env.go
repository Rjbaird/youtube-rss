package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT      string
	REDIS_URL string
	TURSO_URL string
	string
}

func ENV() (*Config, error) {
	godotenv.Load(".env")

	PORT := os.Getenv("PORT")
	if PORT == "" {
		fmt.Println("no PORT environment variable provided")
		fmt.Println("Setting PORT to 3000")
		PORT = "3000"
	}

	REDIS_URL := os.Getenv("REDIS_URL")
	if REDIS_URL == "" {
		log.Fatal("You must set your 'REDIS_URL' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	config := Config{PORT: PORT, REDIS_URL: REDIS_URL}

	return &config, nil
}
