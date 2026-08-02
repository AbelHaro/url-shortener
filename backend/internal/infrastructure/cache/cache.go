package cache

import (
	"fmt"

	"github.com/AbelHaro/url-shortener/backend/internal/config"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	glideConfig "github.com/valkey-io/valkey-glide/go/v2/config"
)

// NewClient connects a standalone GLIDE client to Valkey.
func NewClient(cfg config.CacheConfig) (*glide.Client, error) {
	clientConfig := glideConfig.NewClientConfiguration().WithAddress(&glideConfig.NodeAddress{Host: cfg.Host, Port: cfg.Port})

	client, err := glide.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("create Valkey client: %w", err)
	}

	return client, nil
}
