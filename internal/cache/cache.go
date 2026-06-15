package cache

import (
	"context"
	"time"
)

type Cache interface {
	GetItem(ctx context.Context, key string) ([]byte, error)
	SetItem(ctx context.Context, key string, value []byte, ttl time.Duration) error
}
