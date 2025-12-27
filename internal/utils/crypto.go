// Package utils provides utility functions for the metrics application.
package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"

	"github.com/7StaSH7/gometrics/internal/logger"
	"go.uber.org/zap"
)

var (
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
)

// LoadPublicKey loads an RSA public key from a PEM file.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	publicKey = rsaPub
	return rsaPub, nil
}

// LoadPrivateKey loads an RSA private key from a PEM file.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	privateKey = rsaPriv
	return rsaPriv, nil
}

// Encrypt encrypts data using RSA OAEP with the loaded public key.
func Encrypt(data []byte) ([]byte, error) {
	if publicKey == nil {
		return nil, errors.New("public key not loaded")
	}

	hash := sha256.New()
	encrypted, err := rsa.EncryptOAEP(
		hash,
		rand.Reader,
		publicKey,
		data,
		nil,
	)
	if err != nil {
		logger.Log.Error("encryption failed", zap.Error(err))
		return nil, err
	}

	return encrypted, nil
}

// EncryptToBase64 encrypts data and returns base64 encoded result.
func EncryptToBase64(data []byte) (string, error) {
	encrypted, err := Encrypt(data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts data using RSA OAEP with the loaded private key.
func Decrypt(data []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("private key not loaded")
	}

	hash := sha256.New()
	decrypted, err := rsa.DecryptOAEP(
		hash,
		rand.Reader,
		privateKey,
		data,
		nil,
	)
	if err != nil {
		logger.Log.Error("decryption failed", zap.Error(err))
		return nil, err
	}

	return decrypted, nil
}

// DecryptFromBase64 decrypts base64 encoded data.
func DecryptFromBase64(data string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	return Decrypt(encrypted)
}

// GenerateTestPrivateKey generates a test RSA private key.
func GenerateTestPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// SaveTestKeys saves test keys to temporary files.
func SaveTestKeys(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (string, string, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", "", err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	privateKeyFile := "/tmp/test_private_key.pem"
	if err := os.WriteFile(privateKeyFile, privateKeyPEM, 0600); err != nil {
		return "", "", err
	}

	publicKeyFile := "/tmp/test_public_key.pem"
	if err := os.WriteFile(publicKeyFile, publicKeyPEM, 0644); err != nil {
		return "", "", err
	}

	return privateKeyFile, publicKeyFile, nil
}

// EncryptWithPublicKey encrypts data using the provided public key.
func EncryptWithPublicKey(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	hash := sha256.New()
	encrypted, err := rsa.EncryptOAEP(
		hash,
		rand.Reader,
		publicKey,
		data,
		nil,
	)
	if err != nil {
		logger.Log.Error("encryption failed", zap.Error(err))
		return nil, err
	}

	return encrypted, nil
}
