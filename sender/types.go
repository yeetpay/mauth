package sender

import (
	"context"
	"io"
)

type (
	Sender interface {
		Send(ctx context.Context, address, subject string, txt io.Reader, html io.Reader) error
	}
)
