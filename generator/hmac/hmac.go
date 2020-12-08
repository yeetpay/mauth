/*
	Package hmac generate tokens by signing the email and expiration dates with hmac and sha256. You must provide
	either a 32 or 64 byte hash key. Optionally it can also encrypt the token with AES-128 (if provided a 16
	byte block key) or AES-256 (with a 32 bytes key). Theses keys should be crypto strong random bytes.
	For example to generate a 32 bytes key and encode it to base 64, run this command in a shell:

		head -c 32 /dev/urandom | base64
*/
package hmac

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"log"
	"time"

	"github.com/fdelbos/mauth/generator"
)

type (
	HMAC struct {
		hashKey  []byte
		blockKey []byte
		block    cipher.Block
	}

	Content struct {
		E string
		T int64
	}
)

var (
	ErrKeyNil    = errors.New("key cannot be nil")
	ErrBlockNil  = errors.New("block cannot be nil")
	ErrKeySize   = errors.New("key size must be 32 or 64 bytes")
	ErrBlockSize = errors.New("block size must be 16 bytes for AES-128 or 32 bytes for AES-256")
	ErrIV        = errors.New("failed to generate random iv")
)

func init() {
	gob.Register(Content{})
}

func create(key, block []byte) (*HMAC, error) {
	res := HMAC{}

	if key == nil {
		return nil, ErrKeyNil
	} else if len(key) != 32 && len(key) != 64 {
		return nil, ErrKeySize
	}
	res.hashKey = key

	if block != nil {
		if len(block) != 16 && len(block) != 32 {
			return nil, ErrBlockSize
		}
		res.blockKey = block
		b, err := aes.NewCipher(block)
		if err != nil {
			return nil, err
		}
		res.block = b
	}
	return &res, nil
}

func NewHMAC(key []byte) (*HMAC, error) {
	return create(key, nil)
}

func NewHMACB64(b64Key string) (*HMAC, error) {
	key, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return nil, err
	}
	return NewHMAC(key)
}

func NewHMACWithEncryption(key, block []byte) (*HMAC, error) {
	if block == nil {
		return nil, ErrBlockNil
	}
	return create(key, block)
}

func NewHMACWithEncryptionB64(b64Key, b64Block string) (*HMAC, error) {
	if hash, err := base64.StdEncoding.DecodeString(b64Key); err != nil {
		return nil, err

	} else if block, err := base64.StdEncoding.DecodeString(b64Block); err != nil {
		return nil, err

	} else {
		return NewHMACWithEncryption(hash, block)
	}
}

func (h HMAC) Generate(ctx context.Context, email string, expiration time.Time) (string, error) {
	// 1 - gob marshal
	buff := bytes.Buffer{}
	err := gob.NewEncoder(&buff).Encode(Content{
		E: email,
		T: expiration.Unix()})
	if err != nil {
		return "", err
	}
	content := buff.Bytes()

	// 2 - encrypt if block set
	if h.block != nil {
		if content, err = h.Encrypt(content); err != nil {
			return "", err
		}
	}

	// 3 - sign with HMAC and base64 encode
	return h.Sign(content)
}

func (h HMAC) Validate(ctx context.Context, token string) (string, error) {
	// 1 - check signature and base64 decode
	content, err := h.Unsign(token)
	if err != nil {
		return "", err
	}

	// 2 - decrypt if block is set
	if h.block != nil {
		if content, err = h.Decrypt(content); err != nil {
			log.Print("decrypt")
			return "", err
		}
	}

	// 3 - gob unmarshal
	res := Content{}
	if err := gob.NewDecoder(bytes.NewBuffer(content)).Decode(&res); err != nil {
		return "", generator.ErrInvalid

	}

	// 4 - check expiration time
	expiration := time.Unix(res.T, 0)
	if expiration.Before(time.Now()) {
		return "", generator.ErrInvalid
	}
	return res.E, nil
}
