package valkey

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client for Valkey
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Valkey client from the VALKEY_URL environment variable
func NewClient() (*Client, error) {
	url := os.Getenv("VALKEY_URL")
	if url == "" {
		url = "redis://localhost:6379"
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb}, nil
}

// NewClientWithURL creates a new Valkey client with the given URL
func NewClientWithURL(url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)

	return &Client{rdb: rdb}, nil
}

// GetRedis returns the underlying Redis client for direct operations
func (c *Client) GetRedis() *redis.Client {
	return c.rdb
}

// Close closes the connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping tests the connection
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}
