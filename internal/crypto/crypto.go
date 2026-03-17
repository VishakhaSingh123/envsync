package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// deriveKey creates a 32-byte AES-256 key from a passphrase using SHA-256
func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// Encrypt encrypts plaintext using AES-256-GCM.
// Returns base64-encoded ciphertext (nonce+ciphertext).
// IMPORTANT: The decrypted value is never written to disk.
func Encrypt(plaintext, passphrase string) (string, error) {
	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher creation failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM creation failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce generation failed: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded AES-256-GCM ciphertext.
func Decrypt(encoded, passphrase string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher creation failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM creation failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (wrong key?): %w", err)
	}

	return string(plaintext), nil
}

// EncryptMap encrypts all values in a map using the passphrase.
func EncryptMap(kv map[string]string, passphrase string) (map[string]string, error) {
	result := make(map[string]string, len(kv))
	for k, v := range kv {
		enc, err := Encrypt(v, passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt key %s: %w", k, err)
		}
		result[k] = enc
	}
	return result, nil
}

// DecryptMap decrypts all values in a map.
func DecryptMap(kv map[string]string, passphrase string) (map[string]string, error) {
	result := make(map[string]string, len(kv))
	for k, v := range kv {
		dec, err := Decrypt(v, passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key %s: %w", k, err)
		}
		result[k] = dec
	}
	return result, nil
}

// GetEncryptionKey retrieves the encryption key from an environment variable.
// Never hardcodes or logs the key.
func GetEncryptionKey(envVarName string) (string, error) {
	key := os.Getenv(envVarName)
	if key == "" {
		return "", fmt.Errorf(
			"encryption key not found. Set the environment variable: export %s=$(openssl rand -base64 32)",
			envVarName,
		)
	}
	return key, nil
}

// GenerateKey generates a random 32-byte base64 key suitable for encryption.
func GenerateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
