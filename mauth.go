package mauth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/fdelbos/mauth/templates"

	"github.com/fdelbos/mauth/generator"
	"github.com/fdelbos/mauth/sender"
)

type (
	AddressNormalizer interface {
		Normalize(string) string
	}

	MAuth struct {
		Generator       generator.Generator
		Sender          sender.Sender
		Templates       templates.Templates
		DefaultDuration time.Duration
		DomainWhitelist map[string]interface{}
		DomainBlackList map[string]interface{}
		BaseURL         string
		Param           string
		Normalizer      AddressNormalizer
	}

	preparation struct {
		email      string
		expiration time.Time
		url        string
	}
)

var (
	ErrInvalidBaseURL     = errors.New("only http or https url schemes are supported")
	ErrBlacklistedAddress = errors.New("email address is blacklisted")
)

// NewMAuth creates new MAuth instance with reasonable defaults
func NewMAuth(generator generator.Generator, sender sender.Sender, templates templates.Templates, baseUrl string) (*MAuth, error) {
	checkUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	switch checkUrl.Scheme {
	case "http", "https":
	default:
		return nil, ErrInvalidBaseURL
	}

	return &MAuth{
		Generator:       generator,
		Sender:          sender,
		Templates:       templates,
		DefaultDuration: time.Minute * 20,
		DomainWhitelist: map[string]interface{}{},
		DomainBlackList: map[string]interface{}{"": nil},
		BaseURL:         baseUrl,
		Param:           "mauth_token",
	}, nil
}

func (m MAuth) Send(ctx context.Context, email string) error {
	prep, err := m.prepare(ctx, email, m.DefaultDuration)
	if err != nil {
		return err
	}

	tr, err := m.Templates.Generate(prep.email, prep.url, prep.expiration)
	if err != nil {
		return err
	}

	return m.Sender.Send(ctx, prep.email, tr.Subject, tr.TXT, tr.HTML)
}

func (m MAuth) SendLocalized(ctx context.Context, lang string, email string) error {
	prep, err := m.prepare(ctx, email, m.DefaultDuration)
	if err != nil {
		return err
	}

	tr, err := m.Templates.GenerateForLang(lang, prep.email, prep.url, prep.expiration)
	if err != nil {
		return err
	}

	return m.Sender.Send(ctx, prep.email, tr.Subject, tr.TXT, tr.HTML)
}

func (m MAuth) SendLocalizedFromRequest(ctx context.Context, r *http.Request, email string) error {
	prep, err := m.prepare(ctx, email, m.DefaultDuration)
	if err != nil {
		return err
	}

	tr, err := m.Templates.GenerateForRequest(r, prep.email, prep.url, prep.expiration)
	if err != nil {
		return err
	}

	return m.Sender.Send(ctx, prep.email, tr.Subject, tr.TXT, tr.HTML)
}

func (m MAuth) Validate(ctx context.Context, token string) (string, error) {
	return m.Generator.Validate(ctx, token)
}

func (m MAuth) ValidateRequest(r *http.Request) (string, error) {
	token := r.URL.Query().Get(m.Param)
	if token == "" {
		return "", generator.ErrInvalid
	}
	return m.Generator.Validate(r.Context(), token)
}
