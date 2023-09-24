package web

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
)

type Channel struct {
	Handle      string   `json:"handle"`
	ChannelID   string   `json:"channel_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

func GetDataFromChannel(url string) (*Channel, error) {
	handle := strings.Replace(url, "https://www.youtube.com/", "", -1)

	if !strings.Contains(handle, "@") {
		handle = "@" + handle
	}

	results := Channel{
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
