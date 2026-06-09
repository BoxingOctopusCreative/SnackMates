package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

// Disabled is a cache client used when Valkey is unavailable at startup.
var Disabled = &Client{}

func New(valkeyURL string) (*Client, error) {
	opts, err := redis.ParseURL(valkeyURL)
	if err != nil {
		return nil, fmt.Errorf("parse valkey url: %w", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping valkey: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Available() bool {
	return c != nil && c.rdb != nil
}

func (c *Client) available() bool {
	return c.Available()
}

func (c *Client) Close() error {
	if !c.available() {
		return nil
	}
	return c.rdb.Close()
}

func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if !c.available() {
		return fmt.Errorf("valkey unavailable")
	}
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if !c.available() {
		return "", fmt.Errorf("valkey unavailable")
	}
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if !c.available() {
		return fmt.Errorf("valkey unavailable")
	}
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if !c.available() {
		return fmt.Errorf("valkey unavailable")
	}
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Publish(ctx context.Context, channel, message string) error {
	if !c.available() {
		return fmt.Errorf("valkey unavailable")
	}
	return c.rdb.Publish(ctx, channel, message).Err()
}

func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if !c.available() {
		return nil
	}
	return c.rdb.Subscribe(ctx, channels...)
}
