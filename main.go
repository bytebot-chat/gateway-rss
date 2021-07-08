package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/go-redis/redis/v8"
)

var (
	ctx context.Context
	rdb *redis.Client

	feedURL string
	delay   time.Duration
)

func main() {
	rdb = rdbConnect("127.0.0.1:6379")
	//ctx := context.Background()

	// hardcoded TODO
	feedURL = "http://localhost:8000/flux.xml"
	delay = 3

	feed, err := rss.Fetch(feedURL)
	if err != nil {
		panic(err)
	}

	//	err = feed.Update()
	//if err != nil {
	//	fmt.Println(err)

	for {
		feed.Refresh = time.Now()
		err = feed.Update()
		if err != nil {
			fmt.Println(err)
		}
		printFeedNewItems(feed)

		time.Sleep(time.Second * delay)
	}
	//u2 := uuid.NewV4()
	//fmt.Printf("%s\n", u2)
}

func printFeedNewItems(feed *rss.Feed) {
	for _, i := range feed.Items {
		if !i.Read {
			printItem(i)
			i.Read = true
		}
	}
}

func printItem(item *rss.Item) {
	fmt.Printf("%s: %s (%s)\n", item.Title, item.Summary, item.Link)
}

func rdbConnect(addr string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}
