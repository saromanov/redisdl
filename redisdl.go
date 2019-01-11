package redisdl

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

var errStoreToken = errors.New("unable to store token")

// RedisDL defines main struct of the app
type RedisDL struct {
	client       *redis.Client
	m            sync.Mutex
	key          string
	lockTimeout  time.Duration
	currentToken string
}

// New creates a new app
func New(c *redis.Client, key string) (*RedisDL, error) {
	if _, err := c.Ping().Result(); err != nil {
		return nil, fmt.Errorf("redis is not available: %v", err)
	}

	return &RedisDL{
		client:      c,
		m:           sync.Mutex{},
		key:         key,
		lockTimeout: 5 * time.Second,
	}, nil
}

// Lock provides distributed locking
func (r *RedisDL) Lock(ctx context.Context) error {
	return r.lock(ctx)
}

// Unlock provides unlocking of the store
func (r *RedisDL) Unlock() error {
	r.m.Lock()
	defer r.m.Unlock()
	return r.deleteToken()
}

// resetToken provides removing of current token
func (r *RedisDL) resetToken() {
	r.currentToken = ""
}

func (r *RedisDL) lock(ctx context.Context) error {
	r.m.Lock()
	defer r.m.Unlock()
	token, err := randToken()
	if err != nil {
		return err
	}
	retry := time.NewTimer(r.lockTimeout)
	for {
		if err := r.storeToken(token); err != nil {
			return err
		}
		r.currentToken = token

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retry.C:
		}
	}
}

// storeToken provides store of token
func (r *RedisDL) storeToken(token string) error {
	ok, err := r.client.SetNX(r.key, token, r.lockTimeout).Result()
	if err == redis.Nil {
		err = nil
	}
	if !ok {
		return errStoreToken
	}
	return err

}

// deleteToken provides removing of the token from redis
func (r *RedisDL) deleteToken() error {
	_, err := r.client.Del(r.key).Result()
	return err
}

func randToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
