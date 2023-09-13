package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
)

type RSS struct {
	Handle    string
	ChannelID string
	// Image    string href on image element with id img
	// Title    string innertext on yt-formatted-string element with id text
	// Keywords []string content attribute on meta element with name keywords
}

func main() {
	// setup fiber server
	engine := html.New("./views", ".html")
	server := fiber.New(fiber.Config{Views: engine})

	// define middleware
	server.Use(logger.New())
	server.Use(helmet.New())
	server.Use(cors.New())
	server.Use(recover.New())
	server.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{}, "layouts/main")
	})

	server.Get("/feed", func(c *fiber.Ctx) error {
		url := c.Query("url")
		// if url includes "watch=" then it's a video page
		
		log.Printf("URL: %s", url)
		feed, err := ScrapeRSSFeed(url)
		log.Printf("handle: %s", feed.Handle)
		log.Printf("id: %s", feed.ChannelID)
		if err != nil || feed.ChannelID == "" {
			return c.Render("partials/feed-error", fiber.Map{})
		}
		return c.Render("partials/feed", fiber.Map{
			"Feed": feed,
		})
	})

	log.Fatal(server.Listen(fmt.Sprintf(":%s", "3000")))
}

// RSS from Channel page
func ScrapeRSSFeed(url string) (*RSS, error) {

	handle := strings.Replace(url, "https://www.youtube.com/", "", -1)

	results := RSS{}

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
		return
	})

	c.OnHTML("link[title]", func(e *colly.HTMLElement) {
		if e.Attr("type") != "application/rss+xml" || e.Attr("title") != "RSS" {
			return
		}
		id := strings.Replace(e.Attr("href"), "https://www.youtube.com/feeds/videos.xml?channel_id=", "", -1)
		rss := RSS{
			Handle:    handle,
			ChannelID: id,
		}
		results = rss
	})
	c.Visit(url)
	return &results, nil
}

// RSS from Video page