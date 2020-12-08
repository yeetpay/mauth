package templates

import (
	"errors"
	"net/http"
	"time"
)

var (
	ErrNotFound = errors.New("template not found")
)

type (
	TemplateResult struct {
		HTML    []byte
		TXT     []byte
		Subject string
	}

	Templates interface {
		Generate(email, url string, expiration time.Time) (*TemplateResult, error)
		GenerateForLang(lang string, email, url string, expiration time.Time) (*TemplateResult, error)
		GenerateForRequest(r *http.Request, email, url string, expiration time.Time) (*TemplateResult, error)
	}
)
