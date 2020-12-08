package sender

import (
	"context"
)

type (
	Sender interface {
		Send(ctx context.Context, address, subject string, txt, html []byte) error
	}
)
