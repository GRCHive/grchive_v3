package crypto_utility

import (
	"crypto/hmac"
	"crypto/sha256"
)

func Sha256HMAC(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
