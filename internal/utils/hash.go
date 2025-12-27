// Package utils provides utility functions for the metrics application.
package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/7StaSH7/gometrics/internal/logger"
	"go.uber.org/zap"
)

// GenerateSHA256 generates an HMAC-SHA256 hash of the value using the key.
func GenerateSHA256(value string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSHA256Bytes generates an HMAC-SHA256 hash of the byte value using the key.
func GenerateSHA256Bytes(value []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(value)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySHA256 verifies if the provided hash matches the expected hash.
func VerifySHA256(expectedHash, hash string) bool {
	logger.Log.Debug("Verifying SHA256 hash", zap.String("expectedHash", expectedHash), zap.String("hash", hash))
	return hmac.Equal([]byte(expectedHash), []byte(hash))
}
