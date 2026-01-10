package provider

import (
	"context"
)

type Storage interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
	GetPresignedURL(ctx context.Context, key string, expiry int) (string, error)
}
