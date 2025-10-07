package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/7StaSH7/gometrics/internal/logger"
	"go.uber.org/zap"
)

func GenerateSHA256(value string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifySHA256(expectedHash, hash string) bool {
	logger.Log.Debug("Verifying SHA256 hash", zap.String("expectedHash", expectedHash), zap.String("hash", hash))
	return hmac.Equal([]byte(expectedHash), []byte(hash))
}
