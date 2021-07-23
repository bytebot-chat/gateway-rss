package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/go-redis/redis/v8"
	"github.com/satori/go.uuid"
)

type Message struct {
	*rss.Item
	Metadata Metadata
}

type Metadata struct {
	Source string
	Dest   string
	ID     uuid.UUID
}

func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, m); err != nil {
		return err
	}
	return nil
}

func MessageFromItem(item *rss.Item) *Message {
	msg := new(Message)
	msg.Item = item

	return msg
}

type Feed struct {
	// Cette briganderie
	*rss.Feed
}

func CreateFeed(url string) (Feed, error) {
	feed, err := rss.Fetch(url)
	if err != nil {
		fmt.Println("couldn't load feed")
	}
	return Feed{feed}, err
}

func (f *Feed) PushNewItemsToQueue(rdb *redis.Client,
	inbound string,
	ctx context.Context) error {
	f.Refresh = time.Now()
	err := f.Update()
	if err != nil {
		fmt.Println(err) //TODO proper loggign
	}

	// FIXME This is awfully inefficient
	for _, i := range f.Items {
		if !i.Read {
			msg := MessageFromItem(i)
			msg.Metadata.ID = uuid.Must(uuid.NewV4(), *new(error))
			msg.Metadata.Source = f.UpdateURL
			msg.Metadata.Dest = "gateway-rss"

			stringMsg, _ := json.Marshal(msg)
			rdb.Publish(ctx, inbound, stringMsg)

			i.Read = true
		}
	}
	return nil
}
