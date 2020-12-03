package mauth

import (
	"context"
	"errors"
	"time"

	"github.com/fdelbos/mauth/generator"
	"github.com/fdelbos/mauth/sender"
)

type (
	MAuth struct {
		Generator         generator.Generator
		Sender            sender.Sender
		DefaultExpiration time.Duration
		DomainWhitelist   map[string]interface{}
		DomainBlackList   map[string]interface{}
		CleanEmail        bool
		BaseURL           string
		ParamNam          string
	}
)

var (
	ErrExpiredToken       = errors.New("expired token")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidAddress     = errors.New("invalid email address")
	ErrBlacklistedAddress = errors.New("email address is blacklisted")
)

// NewMAuth creates new MAuth instance with reasonable defaults
func NewMAuth(generator generator.Generator, sender sender.Sender, baseURL string) *MAuth {
	return &MAuth{
		Generator:         generator,
		Sender:            sender,
		DefaultExpiration: time.Minute * 20,
		DomainWhitelist:   map[string]interface{}{},
		DomainBlackList:   map[string]interface{}{},
		CleanEmail:        true,
		BaseURL:           baseURL,
		ParamNam:          "mauth_token",
	}
}

func (m MAuth) Send(ctx context.Context, email string) error {

}

func (m MAuth) SendLocalized(ctx context.Context, email string, languages string) error {

}

func (m MAuth) Validate(ctx context.Context, token string) (string, error) {

}
