package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/bytebot-chat/gateway-rss/model"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"github.com/satori/go.uuid"
)

var (
	ctx context.Context
	rdb *redis.Client

	delay     time.Duration
	feedURL   string
	redisAddr string
	inbound   string
	// I don't think we need to read user messages for a rss gateway?

	err error
)

func init() {
	feedFlag := flag.String("feed", "https://nitter.42l.fr/SwiftOnSecurity/rss", "The rss feed to follow")
	redisFlag := flag.String("redis", "redis:6379", "The redis server's address")
	inboundFlag := flag.String("inbound", "rss-inbound", "The inbound's queue (where the rss items are written)'s name")
	delayFlag := flag.String("delay", "60m", "The delay at which the feed is updated")

	flag.Parse()
	// TODO ENV
	feedURL = *feedFlag
	redisAddr = *redisFlag
	inbound = *inboundFlag
	delay, err = time.ParseDuration(*delayFlag)
	if err != nil {
		log.Warn().
			Err(err).
			Str("delay", *delayFlag).
			Msg("Couldn't parse delay, see https://golang.org/pkg/time/#ParseDuration. Using the default 60m")
	}
}

func main() {
	log.Info().
		Str("feed address", feedURL).
		Str("redis address", redisAddr).
		Str("inbound queue", inbound).
		Dur("update interval", delay).
		Msg("Starting up")

	rdb = rdbConnect(redisAddr)
	ctx = context.Background()

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
}

func pushNewItemsToQueue(feed *rss.Feed) error {
	// FIXME This is awfully inefficient
	for _, i := range feed.Items {
		if !i.Read {
			msg := model.MessageFromItem(i)
			msg.Metadata.ID = uuid.Must(uuid.NewV4(), *new(error))
			msg.Metadata.Source = feedURL
			msg.Metadata.Dest = "gateway-rss"

			stringMsg, _ := json.Marshal(msg)
			rdb.Publish(ctx, inbound, stringMsg)

			i.Read = true
		}
	}
	return nil
}
