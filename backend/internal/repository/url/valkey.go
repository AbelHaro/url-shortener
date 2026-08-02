package url

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/valkey-io/valkey-glide/go/v2/models"
	"github.com/valkey-io/valkey-glide/go/v2/options"
)

var _ Cache = (*ValkeyCache)(nil)

// ValkeyClient is the subset of GLIDE commands needed by the URL cache.
type ValkeyClient interface {
	Get(ctx context.Context, key string) (models.Result[string], error)
	SetWithOptions(ctx context.Context, key, value string, options options.SetOptions) (models.Result[string], error)
	Del(ctx context.Context, keys []string) (int64, error)
}

// ValkeyCache stores JSON-encoded URL records in Valkey.
type ValkeyCache struct {
	client ValkeyClient
	ttl    time.Duration
}

// NewValkeyCache creates a Valkey-backed URL cache with expiring entries.
func NewValkeyCache(client ValkeyClient, ttl time.Duration) (*ValkeyCache, error) {
	if client == nil {
		return nil, errors.New("valkey client is required")
	}
	if ttl <= 0 {
		return nil, errors.New("cache TTL must be positive")
	}
	return &ValkeyCache{client: client, ttl: ttl}, nil
}

func (c *ValkeyCache) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	result, err := c.client.Get(ctx, cacheKey(shortCode))
	if err != nil {
		return nil, fmt.Errorf("get URL from Valkey: %w", err)
	}

	if result.IsNil() {
		return nil, domain.ErrCacheMiss
	}
	var cachedURL domain.URL
	if err := json.Unmarshal([]byte(result.Value()), &cachedURL); err != nil {
		return nil, fmt.Errorf("decode cached URL: %w", err)
	}
	return &cachedURL, nil
}

func (c *ValkeyCache) SetByShortCode(ctx context.Context, shortCode string, url *domain.URL) error {
	payload, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("encode URL for cache: %w", err)
	}

	setOptions := options.NewSetOptions().SetExpiry(options.NewExpiryIn(c.ttl))
	if _, err := c.client.SetWithOptions(ctx, cacheKey(shortCode), string(payload), *setOptions); err != nil {
		return fmt.Errorf("set URL in Valkey: %w", err)
	}
	return nil
}

func (c *ValkeyCache) DeleteByShortCode(ctx context.Context, shortCode string) error {
	_, err := c.client.Del(ctx, []string{cacheKey(shortCode)})
	if err != nil {
		return fmt.Errorf("delete URL from Valkey: %w", err)
	}
	return nil
}

func cacheKey(shortCode string) string {
	return "url:short-code:" + shortCode
}
