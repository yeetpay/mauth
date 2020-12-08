package hmac

import (
	"bytes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"

	"github.com/fdelbos/mauth/generator"
)

func GenerateRandomKey(length int) []byte {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}

func (h HMAC) Encrypt(value []byte) ([]byte, error) {
	if h.block == nil {
		return nil, ErrBlockNil
	}
	iv := GenerateRandomKey(h.block.BlockSize())
	if iv == nil {
		return nil, ErrIV
	}

	// Encrypt it.
	stream := cipher.NewCTR(h.block, iv)
	stream.XORKeyStream(value, value)
	// Return iv + ciphertext.
	return append(iv, value...), nil
}

func (h HMAC) Decrypt(value []byte) ([]byte, error) {
	if h.block == nil {
		return nil, ErrBlockNil
	}
	size := h.block.BlockSize()
	if len(value) <= size {
		return nil, generator.ErrInvalid
	}

	// Extract iv.
	iv := value[:size]
	// Extract ciphertext.
	value = value[size:]
	// Decrypt it.
	stream := cipher.NewCTR(h.block, iv)
	stream.XORKeyStream(value, value)
	return value, nil
}

func (h HMAC) Sign(value []byte) (string, error) {
	valueStr := base64.StdEncoding.EncodeToString(value)
	hash := hmac.New(sha256.New, h.hashKey)
	if _, err := hash.Write([]byte(valueStr)); err != nil {
		return "", err
	}
	sum := hash.Sum(nil)
	valueStr += "#"
	return base64.StdEncoding.EncodeToString(append([]byte(valueStr), sum...)), nil
}

func (h HMAC) Unsign(token string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		log.Print(err)
		return nil, generator.ErrInvalid
	}
	components := bytes.SplitN(b, []byte("#"), 2)
	if len(components) != 2 {
		return nil, generator.ErrInvalid
	}

	mac := hmac.New(sha256.New, h.hashKey)
	_, err = mac.Write(components[0])
	if err != nil {
		return nil, generator.ErrInvalid
	}

	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(components[1], expectedMAC) {
		return nil, generator.ErrInvalid
	}

	return base64.StdEncoding.DecodeString(string(components[0]))
}
