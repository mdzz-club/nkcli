package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

func Encrypt(data []byte, key []byte) ([]byte, error) {
	k := sha256.Sum256(key)
	b, err := aes.NewCipher(k[:])

	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(b)

	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aead.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(data []byte, key []byte) ([]byte, error) {
	k := sha256.Sum256(key)
	b, err := aes.NewCipher(k[:])

	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(b)

	if err != nil {
		return nil, err
	}

	len := aead.NonceSize()
	nonce := data[:len]

	return aead.Open(nil, nonce, data[len:], nil)
}
