package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	gcm     cipher.AEAD
	once    sync.Once
	initErr error
)

// Init loads the encryption key from the ENCRYPTION_KEY environment variable
// and initialises the AES-256-GCM cipher. Call this once at server startup.
//
// The key must be exactly 32 bytes (256 bits) encoded as base64.
// Generate one with: openssl rand -base64 32
func Init() error {
	once.Do(func() {
		keyB64 := os.Getenv("ENCRYPTION_KEY")
		if keyB64 == "" {
			initErr = errors.New("ENCRYPTION_KEY environment variable is not set")
			return
		}

		key, err := base64.StdEncoding.DecodeString(keyB64)
		if err != nil {
			initErr = fmt.Errorf("ENCRYPTION_KEY is not valid base64: %w", err)
			return
		}

		if len(key) != 32 {
			initErr = fmt.Errorf("ENCRYPTION_KEY must be 32 bytes (got %d) — generate with: openssl rand -base64 32", len(key))
			return
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			initErr = fmt.Errorf("failed to create AES cipher: %w", err)
			return
		}

		gcm, err = cipher.NewGCM(block)
		if err != nil {
			initErr = fmt.Errorf("failed to create GCM: %w", err)
			return
		}
	})

	return initErr
}

// Encrypt encrypts plaintext using AES-256-GCM.
//
// Returns a base64-encoded string containing: nonce (12 bytes) || ciphertext+tag.
// This format is safe for storing in text database columns.
func Encrypt(plaintext []byte) (string, error) {
	if gcm == nil {
		return "", errors.New("crypto not initialised — call crypto.Init() first")
	}

	// Generate a random 12-byte nonce (required by GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal prepends the nonce to the ciphertext so we can extract it on decrypt
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode as base64 for safe DB storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext produced by Encrypt.
//
// Expects the format: base64(nonce || ciphertext+tag)
func Decrypt(encoded string) ([]byte, error) {
	if gcm == nil {
		return nil, errors.New("crypto not initialised — call crypto.Init() first")
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong key or corrupted data): %w", err)
	}

	return plaintext, nil
}

// EncryptString is a convenience wrapper that encrypts a string.
func EncryptString(plaintext string) (string, error) {
	return Encrypt([]byte(plaintext))
}

// DecryptString is a convenience wrapper that decrypts to a string.
func DecryptString(encoded string) (string, error) {
	data, err := Decrypt(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
