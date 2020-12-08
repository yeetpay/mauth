package hmac

import (
	"testing"
	"time"

	"github.com/fdelbos/mauth/generator"
	"github.com/stretchr/testify/suite"
)

var (
	b64Key32    = "KKXrtvs26tG3L51nekkHhuzCULqHiSxKu3mXBPFmzgk="
	b64Key64    = "kCawJLp4y/mpV+67Evb9WZmSvSE0JOR9bw+Z+O6drflGzxh0datdLDaTmiocdRyVv2H7kQgUmTW4UO8pp8zpDg=="
	b64Block16  = "0m1xi6PHehHtjaS1KjnNHg=="
	b64Block32  = "o1xYpm5sXMJKjm/q+uFOp1Ft5wp261zAgYPdPPp9kAw="
	b64TooSmall = "VIXqIHpLYpc="
	email       = "test@example.com"
)

type (
	HMACSuite struct {
		suite.Suite
	}
)

func TestHMACSuite(t *testing.T) {
	suite.Run(t, &HMACSuite{})
}

func (s *HMACSuite) TestEncodingDecoding() {
	cases := []struct {
		msg     string
		key     string
		block   *string
		errInit error
	}{
		{"32 key, no block, no error", b64Key32, nil, nil},
		{"32 key, 12 block, no error", b64Key32, &b64Block16, nil},
		{"32 key, 32 block, no error", b64Key32, &b64Block32, nil},
		{"32 key, invalid block, ErrBlockSize", b64Key32, &b64TooSmall, ErrBlockSize},
		//
		{"64 key, no block, no error", b64Key64, nil, nil},
		{"64 key, 16 block, no error", b64Key64, &b64Block16, nil},
		{"64 key, 32 block, no error", b64Key64, &b64Block32, nil},
		{"64 key, invalid block, ErrBlockSize", b64Key64, &b64TooSmall, ErrBlockSize},
		//
		{"invalid key, no block, ErrKeySize", b64TooSmall, nil, ErrKeySize},
		{"invalid key, 16 block, ErrKeySize", b64TooSmall, &b64Block16, ErrKeySize},
		{"invalid key, 32 block, ErrKeySize", b64TooSmall, &b64Block32, ErrKeySize},
		{"invalid key, invalid block, ErrKeySize", b64TooSmall, &b64TooSmall, ErrKeySize},
	}

	for _, c := range cases {
		var hmac *HMAC
		var err error
		if c.block != nil {
			hmac, err = NewHMACWithEncryptionB64(c.key, *c.block)
		} else {
			hmac, err = NewHMACB64(c.key)
		}
		s.Require().Equal(c.errInit, err, c.msg)
		if err != nil {
			continue
		}
		token, err := hmac.Generate(nil, email, time.Now().Add(1*time.Minute))
		s.Require().Nil(err, c.msg)

		res, err := hmac.Validate(nil, token)
		s.Require().Nil(err, c.msg)
		s.Require().Equal(email, res)
	}
}

func (s *HMACSuite) TestExpired() {
	hmac, err := NewHMACWithEncryptionB64(b64Key64, b64Block32)
	s.Require().Nil(err)

	token, err := hmac.Generate(nil, email, time.Now().Add(-time.Minute))
	s.Require().Nil(err)

	res, err := hmac.Validate(nil, token)
	s.Require().Equal(generator.ErrInvalid, err)
	s.Require().Equal("", res)
}
