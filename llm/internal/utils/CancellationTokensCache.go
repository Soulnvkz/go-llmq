package utils

import (
	"context"
	"sync"
	"time"
)

type CancelToken struct {
	Ctx    *context.Context
	Cancel *context.CancelFunc
}

type CancellationTokensCache struct {
	m        map[string]*CancelToken
	lifetime time.Duration

	created []int64
	keys    []string

	mu sync.Mutex
}

func NewCancellationTokensCache(ctx context.Context, lifetime time.Duration, tick time.Duration) *CancellationTokensCache {
	ticker := time.NewTicker(tick)

	c := &CancellationTokensCache{
		m:        make(map[string]*CancelToken),
		lifetime: lifetime,
		created:  make([]int64, 0, 1000),
		keys:     make([]string, 0, 1000),
		mu:       sync.Mutex{},
	}

	go func(cache *CancellationTokensCache) {
	loop:
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				break loop
			case t := <-ticker.C:
				_ = t
				c.mu.Lock()

				i := 0
				for ; i < len(c.created); i++ {
					diff := t.Unix() - c.created[i]
					if diff < int64(c.lifetime) {
						break
					}

					delete(c.m, c.keys[i])
				}
				c.created = c.created[i:]
				c.keys = c.keys[i:]

				c.mu.Unlock()
			}
		}
	}(c)

	return c
}

func (c *CancellationTokensCache) Put(key string, token *CancelToken) {
	c.mu.Lock()
	c.m[key] = token
	c.created = append(c.created, time.Now().Unix())
	c.keys = append(c.keys, key)
	c.mu.Unlock()
}

func (c *CancellationTokensCache) Get(key string) (v *CancelToken, ok bool) {
	c.mu.Lock()
	v, ok = c.m[key]
	c.mu.Unlock()
	return
}
