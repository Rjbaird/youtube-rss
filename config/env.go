package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT       string
	REDIS_URL  string
	TURSO_URL  string
	TURSO_AUTH string
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

	TURSO_URL := os.Getenv("TURSO_URL")
	if TURSO_URL == "" {
		log.Fatal("You must set your 'TURSO_URL' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	TURSO_AUTH := os.Getenv("TURSO_AUTH")
	if TURSO_AUTH == "" {
		log.Fatal("You must set your 'TURSO_AUTH' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	config := Config{PORT: PORT, REDIS_URL: REDIS_URL, TURSO_URL: TURSO_URL, TURSO_AUTH: TURSO_AUTH}

	return &config, nil
}
