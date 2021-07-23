package main

import (
	"context"
	"flag"
	"time"

	"github.com/bytebot-chat/gateway-rss/model"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

var (
	ctx   context.Context
	rdb   *redis.Client
	feeds []model.Feed

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

	feeds = make([]model.Feed, 0)
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

	feed, err := model.CreateFeed(feedURL)
	if err != nil {
		panic(err)
	}

	feeds = append(feeds, feed)

	for {
		for _, f := range feeds {
			go f.PushNewItemsToQueue(rdb, inbound, ctx)
		}
		time.Sleep(time.Second * delay)
	}
}
