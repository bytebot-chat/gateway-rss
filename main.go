package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/bytebot-chat/gateway-rss/model"
	"github.com/go-redis/redis/v8"
	"github.com/satori/go.uuid"
)

var (
	ctx context.Context
	rdb *redis.Client

	delay time.Duration

	feedURL   = "http://localhost:8000/flux.xml"
	redisAddr = "127.0.0.1:6379"
	inbound   = "rss-inbound"
	// I don't think we need an outbound queue for a rss gateway?
)

func main() {
	rdb = rdbConnect("127.0.0.1:6379")
	ctx = context.Background()

	// hardcoded TODO
	delay = 3 //TODO

	feed, err := rss.Fetch(feedURL)
	if err != nil {
		panic(err)
	}

	for {
		feed.Refresh = time.Now()
		err = feed.Update()
		if err != nil {
			fmt.Println(err)
		}
		pushNewItemsToQueue(feed)

		time.Sleep(time.Second * delay)
	}
	//u2 := uuid.NewV4()
	//fmt.Printf("%s\n", u2)
}

func pushNewItemsToQueue(feed *rss.Feed) error {
	// This is awfully inefficient
	for _, i := range feed.Items {
		if !i.Read {
			msg := model.MessageFromItem(i)
			msg.Metadata.ID = uuid.Must(uuid.NewV4(), *new(error))
			msg.Metadata.Source = feedURL
			msg.Metadata.Dest = "gateway-rss"

			stringMsg, _ := json.Marshal(msg)
			rdb.Publish(ctx, inbound, stringMsg)

			// TODO publish to pubsub
			i.Read = true
		}
	}
	return nil
}

// Utils
func printFeedNewItems(feed *rss.Feed) {
	for _, i := range feed.Items {
		if !i.Read {
			printItem(i)
			i.Read = true
		}
	}
}

func printItem(item *rss.Item) {
	//fmt.Printf("%s: %s (%s)\n", item.Title, item.Summary, item.Link)
	m, _ := json.Marshal(item)
	fmt.Println(string(m))
}

func rdbConnect(addr string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}
