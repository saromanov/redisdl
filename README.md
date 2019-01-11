# redisdl
[![Build Status](https://travis-ci.org/saromanov/redisdl.svg?branch=master)](https://travis-ci.org/saromanov/redisdl)
[![Go Report Card](https://goreportcard.com/badge/github.com/saromanov/redisdl)](https://goreportcard.com/report/github.com/saromanov/redisdl)
[![Coverage Status](https://coveralls.io/repos/github/saromanov/redisdl/badge.svg?branch=master)](https://coveralls.io/github/saromanov/redisdl?branch=master)

Implementation of distributed locking over Redis

## Example
```go
package main

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/saromanov/redisdl"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()
	dl, err := redisdl.New(&redisdl.Options{
		Client:     client,
		Key:        "fun.lock",
		RetryCount: 3,
	})
	if err != nil {
		panic(err)
	}
	dl.Lock(context.Background())
	defer dl.Unlock()
}
```