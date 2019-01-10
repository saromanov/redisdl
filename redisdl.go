package redisdl

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-redis/redis"
)

// RedisDL defines main struct of the app
type RedisDL struct {
	client *redis.Client
	m      sync.Mutex
}

// New creates a new app
func New(c *redis.Client, fileName string) (*RedisDL, error) {
	if err := c.Ping(); err != nil {
		return nil, fmt.Errorf("redis is not available: %v", err)
	}

	return &RedisDL{
		client: c,
		m:      sync.Mutex{},
	}, nil
}

// Lock provides distributed locking
func (r *RedisDL) Lock(ctx context.Context) error {
	return r.lock(ctx)
}

func (r *RedisDL) lock(ctx context.Context) error {
	r.m.Lock()
	defer r.m.Unlock()
	return nil
}
