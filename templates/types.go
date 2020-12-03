package templates

import (
	"io"
	"time"
)

type (
	Templates interface {
		GenerateTXT(lang, email, token string, expiration time.Time) (io.Reader, error)
		GenerateHTML(lang, email, token string, expiration time.Time) (io.Reader, error)
	}
)
