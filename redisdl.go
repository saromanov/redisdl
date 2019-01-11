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

const defaultKey = "redisdl.key"

var errStoreToken = errors.New("unable to store token")

// Options defines struct with options to the app
type Options struct {
	LockTimeout time.Duration
	Key         string
	RetryCount  uint
	Client      *redis.Client
}

func (o *Options) setDefault() {
	if o.Key == "" {
		o.Key = defaultKey
	}
	if o.RetryCount == 0 {
		o.RetryCount = 5
	}
}

// RedisDL defines main struct of the app
type RedisDL struct {
	client       *redis.Client
	m            sync.Mutex
	currentToken string
	opt          *Options
}

// New creates a new app
func New(opt *Options) (*RedisDL, error) {
	opt.setDefault()
	if _, err := opt.Client.Ping().Result(); err != nil {
		return nil, fmt.Errorf("redis is not available: %v", err)
	}

	return &RedisDL{
		client: opt.Client,
		m:      sync.Mutex{},
		opt:    opt,
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

// GetToken returns currrent token
func (r *RedisDL) GetToken() string {
	return r.currentToken
}

// resetToken provides removing of current token
func (r *RedisDL) resetToken() {
	r.currentToken = ""
}

// lock provides trying to gen new token
func (r *RedisDL) lock(ctx context.Context) error {
	r.m.Lock()
	defer r.m.Unlock()
	token, err := randToken()
	if err != nil {
		return err
	}
	retry := time.NewTimer(r.opt.LockTimeout)
	atts := r.opt.RetryCount + 1
	for {
		if err := r.storeToken(token); err == nil {
			r.currentToken = token
			return nil
		}
		if atts--; atts <= 0 {
			return fmt.Errorf("unable to generate token")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retry.C:
		}
	}
}

// storeToken provides store of token
func (r *RedisDL) storeToken(token string) error {
	ok, err := r.client.SetNX(r.opt.Key, token, r.opt.LockTimeout).Result()
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
	_, err := r.client.Del(r.opt.Key).Result()
	return err
}

func randToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
