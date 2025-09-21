package kubeflake

import (
	"bytes"
	"errors"
)

const (
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

var (
	base62Bytes    = []byte(base62Chars)
	base64Bytes    = []byte(base64Chars)
	ErrInvalidBase = errors.New("invalid base")
)

type Base62Converter struct{}

var _ BaseConverter = (*Base62Converter)(nil)

// EncodeBase62 converts an uint64 to a base62-encoded string.
func (Base62Converter) Encode(n uint64) string {
	if n == 0 {
		return "0"
	}
	result := make([]byte, 0)
	for n > 0 {
		remainder := n % 62
		result = append([]byte{base62Chars[remainder]}, result...)
		n = n / 62
	}
	return string(result)
}

// DecodeBase62 converts a base62-encoded string to an uint64
func (Base62Converter) Decode(s string) (uint64, error) {
	var result uint64
	for i := 0; i < len(s); i++ {
		char := s[i]
		index := bytes.IndexByte(base62Bytes, char)
		if index == -1 {
			return 0, ErrInvalidBase
		}
		result = result*62 + uint64(index)
	}
	return result, nil
}

type Base64Converter struct{}

var _ BaseConverter = (*Base64Converter)(nil)

// EncodeBase64 converts an uint64 to a base64-encoded string.
func (Base64Converter) Encode(n uint64) string {
	if n == 0 {
		return "0"
	}
	result := make([]byte, 0)
	for n > 0 {
		remainder := n % 64
		result = append([]byte{base64Chars[remainder]}, result...)
		n = n / 64
	}
	return string(result)
}

// DecodeBase64 converts a base64-encoded string to an uint64
func (Base64Converter) Decode(s string) (uint64, error) {
	var result uint64
	for i := 0; i < len(s); i++ {
		char := s[i]
		index := bytes.IndexByte(base64Bytes, char)
		if index == -1 {
			return 0, ErrInvalidBase
		}
		result = result*64 + uint64(index)
	}
	return result, nil
}

type BaseConverter interface {
	Encode(n uint64) string
	Decode(s string) (uint64, error)
}
