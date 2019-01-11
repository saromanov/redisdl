package redisdl

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

func TestNew(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()
	_, err := New(&Options{
		Client: client,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidNew(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:123",
	})
	defer client.Close()
	_, err := New(&Options{
		Client: client,
	})
	if err == nil {
		t.Fatal("should't start redis with invalid port")
	}
}

func TestLock(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()
	d, err := New(&Options{
		Client:      client,
		Key:         "mock2.key",
		LockTimeout: 3 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = d.Lock(context.Background())
	defer d.Unlock()
	token := d.GetToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("unable to get token")
	}
}
