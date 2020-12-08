package mauth

import (
	"context"
	"net/url"
	"strings"
	"time"
)

func (m MAuth) prepare(ctx context.Context, email string, duration time.Duration) (*preparation, error) {
	if err := m.checkIfAuthorized(email); err != nil {
		return nil, err
	}
	if m.Normalizer != nil {
		email = m.Normalizer.Normalize(email)
	}
	expiration := time.Now().Add(duration)

	token, err := m.Generator.Generate(ctx, email, expiration)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(m.BaseURL)
	if err != nil {
		return nil, err
	}
	q := baseURL.Query()
	q.Set(m.Param, token)
	baseURL.RawQuery = q.Encode()

	return &preparation{
		email:      email,
		expiration: expiration,
		url:        baseURL.String(),
	}, nil
}

func (m MAuth) checkIfAuthorized(email string) error {
	domain := getDomain(email)

	if len(m.DomainBlackList) != 0 {
		if _, ok := m.DomainBlackList[domain]; ok {
			return ErrBlacklistedAddress
		}
	}

	if len(m.DomainWhitelist) != 0 {
		if _, ok := m.DomainBlackList[domain]; !ok {
			return ErrBlacklistedAddress
		}
	}

	return nil
}

func getDomain(email string) string {
	prepared := strings.TrimSpace(email)
	prepared = strings.TrimRight(prepared, ".")

	parts := strings.Split(prepared, "@")
	if len(parts) != 2 {
		return ""
	}

	domain := strings.ToLower(parts[1]) // Domain names are case-insensitive (RFC 4343)
	return domain
}
