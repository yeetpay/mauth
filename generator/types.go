package generator

import (
	"context"
	"time"
)

type (
	Generator interface {
		Generate(ctx context.Context, email string, expiration time.Time) (string, error)
		Validate(ctx context.Context, token string) (string, error)
	}
)
