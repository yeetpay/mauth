package generator

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid = errors.New("token expired, not found or invalid")
)

type (
	Generator interface {
		Generate(ctx context.Context, email string, expiration time.Time) (string, error)
		Validate(ctx context.Context, token string) (string, error)
	}
)
