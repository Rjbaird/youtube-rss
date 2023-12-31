package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bairrya/youtube-rss/db"
	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/gomodule/redigo/redis"
)

func main() {
	// ctx := context.Background()

	// setup server
	engine := html.New("./views", ".html")
	server := fiber.New(fiber.Config{Views: engine})

	// setup database & search
	r, err := db.RedisStackConnect()
	if err != nil {
		log.Fatalf("Error connecting to redis: %s", err)
	}
	defer r.Client.Close()

	// define middleware
	server.Use(logger.New())
	server.Use(helmet.New())
	server.Use(cors.New())
	server.Use(recover.New())
	server.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("x-forwarded-for")
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Render("limit", fiber.Map{})
		},
	}))

	server.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{}, "layouts/main")
	})

	server.Get("/feed", func(c *fiber.Ctx) error {
		url := c.Query("url")
		if strings.Contains(url, "watch?v=") {
			return c.Render("partials/channel-error", fiber.Map{})
		}
		// TODO: add switch for shorts, videos, and other non-channel urls

		handle := strings.Replace(url, "https://www.youtube.com/", "", -1)
		// TODO: check for /videos and other paths and remove

		val, err := redis.Bytes(r.ReJSON.JSONGet(handle, "."))
		if err != nil {
			feed, err := getDataFromChannel(url)
			if err != nil || feed.Title == "" {
				log.Printf("Something went wrong getting channel data for %s: %s", handle, err)
				return c.Render("partials/feed-error", fiber.Map{})
			}

			// TODO: marshal to JSON
			jsonFeed, _ := json.Marshal(feed)

			res, err := r.ReJSON.JSONSet(feed.Handle, ".", jsonFeed)
			if err != nil || res.(string) != "OK" {
				log.Printf("Failed to save %s to redis: %s", feed.Handle, err)
			}

			if res.(string) == "OK" {
				log.Printf("Saved %s to redis", feed.Handle)
			}

			return c.Render("partials/feed", fiber.Map{
				"Feed":     feed,
				"Keywords": feed.Keywords,
			})
		}
		readChannel := db.Channel{}
		err = json.Unmarshal(val, &readChannel)
		if err != nil {
			log.Printf("Failed to JSON Unmarshal")
		}
		log.Printf("found: %s", readChannel.Handle)
		return c.Render("partials/feed", fiber.Map{
			"Feed":     readChannel,
			"Keywords": readChannel.Keywords,
		})
	})

	log.Fatal(server.Listen(fmt.Sprintf(":%s", "3000")))
}

func getDataFromChannel(url string) (*db.Channel, error) {
	handle := strings.Replace(url, "https://www.youtube.com/", "", -1)

	if !strings.Contains(handle, "@") {
		handle = "@" + handle
	}

	results := db.Channel{
		Handle: handle,
	}
	tags := []string{}

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == 404 {
			log.Println("Page not found", r.Request.URL)
			return
		}
		log.Println("Something went wrong getting rss feed:", err)
	})

	c.OnHTML("meta", func(e *colly.HTMLElement) {
		switch e.Attr("property") {
		case "og:title":
			results.Title = e.Attr("content")
		case "og:url":
			results.ChannelID = strings.Replace(e.Attr("content"), "https://www.youtube.com/channel/", "", -1)
		case "description":
			results.Description = e.Attr("content")
		case "og:video:tag":
			tags = append(tags, e.Attr("content"))
		}

	})

	c.Visit(url)

	results.Keywords = tags
	return &results, nil
}
