package url

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/valkey-io/valkey-glide/go/v2/models"
	"github.com/valkey-io/valkey-glide/go/v2/options"
)

type fakeValkeyClient struct {
	values      map[string]string
	getErr      error
	setErr      error
	deleteErr   error
	setOptions  options.SetOptions
	deletedKeys []string
}

func newFakeValkeyClient() *fakeValkeyClient {
	return &fakeValkeyClient{values: make(map[string]string)}
}

func (client *fakeValkeyClient) Get(_ context.Context, key string) (models.Result[string], error) {
	if client.getErr != nil {
		return models.CreateNilStringResult(), client.getErr
	}
	value, ok := client.values[key]
	if !ok {
		return models.CreateNilStringResult(), nil
	}
	return models.CreateStringResult(value), nil
}

func (client *fakeValkeyClient) SetWithOptions(
	_ context.Context,
	key string,
	value string,
	setOptions options.SetOptions,
) (models.Result[string], error) {
	if client.setErr != nil {
		return models.CreateNilStringResult(), client.setErr
	}
	client.values[key] = value
	client.setOptions = setOptions
	return models.CreateStringResult("OK"), nil
}

func (client *fakeValkeyClient) Del(_ context.Context, keys []string) (int64, error) {
	if client.deleteErr != nil {
		return 0, client.deleteErr
	}
	var deleted int64
	for _, key := range keys {
		if _, ok := client.values[key]; ok {
			deleted++
		}
		delete(client.values, key)
	}
	client.deletedKeys = append(client.deletedKeys, keys...)
	return deleted, nil
}

func TestNewValkeyCacheValidation(t *testing.T) {
	tests := []struct {
		name    string
		client  ValkeyClient
		ttl     time.Duration
		wantErr bool
	}{
		{name: "valid", client: newFakeValkeyClient(), ttl: time.Hour},
		{name: "nil client", ttl: time.Hour, wantErr: true},
		{name: "zero TTL", client: newFakeValkeyClient(), wantErr: true},
		{name: "negative TTL", client: newFakeValkeyClient(), ttl: -time.Second, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewValkeyCache(tt.client, tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewValkeyCache() error = %v, wantErr %t", err, tt.wantErr)
			}
			if !tt.wantErr && cache == nil {
				t.Fatal("NewValkeyCache() returned nil cache")
			}
		})
	}
}

func TestValkeyCacheRoundTrip(t *testing.T) {
	client := newFakeValkeyClient()
	cache, err := NewValkeyCache(client, time.Hour)
	if err != nil {
		t.Fatalf("NewValkeyCache() error = %v", err)
	}
	storedURL := &domain.URL{
		ID:          uuid.New(),
		OriginalURL: "https://example.com/path",
		ShortCode:   "abc123",
		UserID:      uuid.New(),
	}

	if err := cache.SetByShortCode(context.Background(), storedURL.ShortCode, storedURL); err != nil {
		t.Fatalf("SetByShortCode() error = %v", err)
	}
	args, err := client.setOptions.ToArgs()
	if err != nil {
		t.Fatalf("SetOptions.ToArgs() error = %v", err)
	}
	if !reflect.DeepEqual(args, []string{"EX", "3600"}) {
		t.Errorf("SET options = %v, want [EX 3600]", args)
	}

	foundURL, err := cache.GetByShortCode(context.Background(), storedURL.ShortCode)
	if err != nil {
		t.Fatalf("GetByShortCode() error = %v", err)
	}
	if !reflect.DeepEqual(foundURL, storedURL) {
		t.Errorf("GetByShortCode() = %+v, want %+v", foundURL, storedURL)
	}

	if err := cache.DeleteByShortCode(context.Background(), storedURL.ShortCode); err != nil {
		t.Fatalf("DeleteByShortCode() error = %v", err)
	}
	_, err = cache.GetByShortCode(context.Background(), storedURL.ShortCode)
	if !errors.Is(err, domain.ErrCacheMiss) {
		t.Errorf("GetByShortCode() after delete error = %v, want %v", err, domain.ErrCacheMiss)
	}
}

func TestValkeyCacheErrors(t *testing.T) {
	t.Run("get command", func(t *testing.T) {
		client := newFakeValkeyClient()
		client.getErr = errors.New("GET failed")
		cache, err := NewValkeyCache(client, time.Hour)
		if err != nil {
			t.Fatalf("NewValkeyCache() error = %v", err)
		}
		if _, err := cache.GetByShortCode(context.Background(), "abc"); err == nil {
			t.Fatal("GetByShortCode() error = nil, want error")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		client := newFakeValkeyClient()
		client.values[cacheKey("abc")] = "not-json"
		cache, err := NewValkeyCache(client, time.Hour)
		if err != nil {
			t.Fatalf("NewValkeyCache() error = %v", err)
		}
		if _, err := cache.GetByShortCode(context.Background(), "abc"); err == nil {
			t.Fatal("GetByShortCode() error = nil, want error")
		}
	})

	t.Run("set command", func(t *testing.T) {
		client := newFakeValkeyClient()
		client.setErr = errors.New("SET failed")
		cache, err := NewValkeyCache(client, time.Hour)
		if err != nil {
			t.Fatalf("NewValkeyCache() error = %v", err)
		}
		if err := cache.SetByShortCode(context.Background(), "abc", &domain.URL{}); err == nil {
			t.Fatal("SetByShortCode() error = nil, want error")
		}
	})

	t.Run("delete command", func(t *testing.T) {
		client := newFakeValkeyClient()
		client.deleteErr = errors.New("DEL failed")
		cache, err := NewValkeyCache(client, time.Hour)
		if err != nil {
			t.Fatalf("NewValkeyCache() error = %v", err)
		}
		if err := cache.DeleteByShortCode(context.Background(), "abc"); err == nil {
			t.Fatal("DeleteByShortCode() error = nil, want error")
		}
	})
}

func TestCacheKeyNamespacesShortCodes(t *testing.T) {
	if got, want := cacheKey("abc123"), "url:short-code:abc123"; got != want {
		t.Errorf("cacheKey() = %q, want %q", got, want)
	}
}
