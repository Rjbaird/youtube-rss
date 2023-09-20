package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"

	"github.com/bairrya/youtube-rss/db"
)

type RSS struct {
	Handle      string
	ChannelID   string
	Title       string
	Description string
	Image       string
	Keywords    []string
	// TODO: add timestamp?
}

func main() {
	fmt.Printf("Starting server...\n")
	// ctx := context.Background()

	// setup fiber server
	engine := html.New("./views", ".html")
	server := fiber.New(fiber.Config{Views: engine})

	// setup database
	db, err := db.RedisConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// define middleware
	server.Use(logger.New())
	server.Use(helmet.New())
	server.Use(cors.New())
	server.Use(recover.New())
	server.Use(limiter.New(limiter.Config{
		// Next: func(c *fiber.Ctx) bool {
		// 	return c.IP() == "127.0.0.1"
		// },
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
		keys, _, err := db.Scan(0, "*", 9).Result()
		if err != nil {
			log.Printf("Something went wrong getting recent keys: %s", err)
		}
		vals, err := db.MGet(keys...).Result()
		if err != nil {
			log.Printf("Something went wrong getting values: %s", err)
		}
		feeds := []RSS{}
		for _, val := range vals {
			feeds = append(feeds, feedToStruct(val.(string)))
		}
		return c.Render("index", fiber.Map{
			"Feeds": feeds,
		}, "layouts/main")
	})

	server.Get("/feed", func(c *fiber.Ctx) error {
		url := c.Query("url")
		if strings.Contains(url, "watch?v=") {
			return c.Render("partials/channel-error", fiber.Map{})
		}
		// TODO: add switch for shorts, videos, and other non-channel urls

		handle := strings.Replace(url, "https://www.youtube.com/", "", -1)
		// TODO: check for /videos and other paths and remove

		val, err := db.Get(handle).Result()
		if err != nil {
			feed, err := getDataFromChannel(url)
			if err != nil || feed.Title == "" {
				return c.Render("partials/feed-error", fiber.Map{})
			}

			log.Printf("adding: %s", feed.Handle)

			err = db.Set(feed.Handle, channelToValue(*feed), 0).Err()
			if err != nil {
				log.Printf("Something went wrong setting %s: %s", feed.Handle, err)
			}

			return c.Render("partials/feed", fiber.Map{
				"Feed":     feed,
				"Keywords": feed.Keywords,
			})
		}

		feed := feedToStruct(val)
		log.Printf("found: %s", feed.Handle)
		return c.Render("partials/feed", fiber.Map{
			"Feed":     feed,
			"Keywords": feed.Keywords,
		})
	})

	log.Fatal(server.Listen(fmt.Sprintf(":%s", "3000")))
}

func getDataFromChannel(url string) (*RSS, error) {
	handle := strings.Replace(url, "https://www.youtube.com/", "", -1)

	if !strings.Contains(handle, "@") {
		handle = "@" + handle
	}

	results := RSS{
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
		case "og:image":
			results.Image = e.Attr("content")
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

func keywordsToString(keywords []string) string {
	return strings.Join(keywords[:], ",")
}

func keywordsToSlice(keywords string) []string {
	return strings.Split(keywords, ",")
}

func channelToValue(channel RSS) string {
	return fmt.Sprintf("%s::%s::%s::%s::%s::%s", channel.Handle, channel.ChannelID, channel.Title, channel.Description, channel.Image, channel.Keywords)
}

func feedToStruct(feed string) RSS {
	channelSlice := strings.Split(feed, "::")
	kewyords := strings.Split(channelSlice[5], ",")
	return RSS{
		Handle:      channelSlice[0],
		ChannelID:   channelSlice[1],
		Title:       channelSlice[2],
		Description: channelSlice[3],
		Image:       channelSlice[4],
		Keywords:    kewyords,
	}
}
