package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

type testKeys struct {
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
	privateKeyFile string
	publicKeyFile  string
}

func generateTestKeys(t *testing.T) testKeys {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	publicKey := &privateKey.PublicKey

	tempDir := t.TempDir()

	privateKeyFile := filepath.Join(tempDir, "private_key.pem")
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	if err := os.WriteFile(privateKeyFile, privateKeyPEM, 0600); err != nil {
		t.Fatalf("failed to write private key file: %v", err)
	}

	publicKeyFile := filepath.Join(tempDir, "public_key.pem")
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if err := os.WriteFile(publicKeyFile, publicKeyPEM, 0644); err != nil {
		t.Fatalf("failed to write public key file: %v", err)
	}

	return testKeys{
		privateKey:     privateKey,
		publicKey:      publicKey,
		privateKeyFile: privateKeyFile,
		publicKeyFile:  publicKeyFile,
	}
}

func TestCrypto(t *testing.T) {
	keys := generateTestKeys(t)

	loadedPubKey, err := LoadPublicKey(keys.publicKeyFile)
	if err != nil {
		t.Fatalf("failed to load public key: %v", err)
	}

	loadedPrivKey, err := LoadPrivateKey(keys.privateKeyFile)
	if err != nil {
		t.Fatalf("failed to load private key: %v", err)
	}

	t.Run("LoadPublicKey", func(t *testing.T) {
		if loadedPubKey == nil {
			t.Error("LoadPublicKey() returned nil key")
		}
	})

	t.Run("LoadPublicKey_FileNotFound", func(t *testing.T) {
		_, err := LoadPublicKey(filepath.Join(keys.publicKeyFile, "..", "nonexistent.pem"))
		if err == nil {
			t.Error("LoadPublicKey() expected error for non-existent file, got nil")
		}
	})

	t.Run("LoadPrivateKey", func(t *testing.T) {
		if loadedPrivKey == nil {
			t.Error("LoadPrivateKey() returned nil key")
		}
	})

	t.Run("LoadPrivateKey_FileNotFound", func(t *testing.T) {
		_, err := LoadPrivateKey(filepath.Join(keys.privateKeyFile, "..", "nonexistent.pem"))
		if err == nil {
			t.Error("LoadPrivateKey() expected error for non-existent file, got nil")
		}
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		if loadedPubKey == nil {
			t.Fatal("loadedPubKey is nil")
		}
		if loadedPrivKey == nil {
			t.Fatal("loadedPrivKey is nil")
		}

		hash := sha256.New()

		testData := []byte("test message for encryption")
		encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, loadedPubKey, testData, nil)
		if err != nil {
			t.Errorf("Encrypt() error = %v", err)
			return
		}

		if len(encrypted) == 0 {
			t.Error("Encrypt() returned empty data")
			return
		}

		decrypted, err := rsa.DecryptOAEP(hash, rand.Reader, loadedPrivKey, encrypted, nil)
		if err != nil {
			t.Errorf("Decrypt() error = %v", err)
			return
		}

		if string(decrypted) != string(testData) {
			t.Errorf("Decrypt() = %v, want %v", string(decrypted), string(testData))
		}
	})

	t.Run("EncryptDecrypt_LargeData", func(t *testing.T) {
		hash := sha256.New()

		testData := make([]byte, 190)
		for i := range testData {
			testData[i] = byte(i % 256)
		}

		encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, loadedPubKey, testData, nil)
		if err != nil {
			t.Errorf("Encrypt() error for large data = %v", err)
			return
		}

		decrypted, err := rsa.DecryptOAEP(hash, rand.Reader, loadedPrivKey, encrypted, nil)
		if err != nil {
			t.Errorf("Decrypt() error for large data = %v", err)
			return
		}

		if string(decrypted) != string(testData) {
			t.Error("Decrypt() large data mismatch")
		}
	})
}

func TestGlobalCryptoFunctions(t *testing.T) {
	keys := generateTestKeys(t)

	t.Run("EncryptDecrypt_GlobalFunctions", func(t *testing.T) {
		if _, err := LoadPublicKey(keys.publicKeyFile); err != nil {
			t.Fatalf("failed to load public key: %v", err)
		}
		if _, err := LoadPrivateKey(keys.privateKeyFile); err != nil {
			t.Fatalf("failed to load private key: %v", err)
		}

		testData := []byte("test message for global functions")
		encrypted, err := Encrypt(testData)
		if err != nil {
			t.Errorf("Encrypt() error = %v", err)
			return
		}

		if len(encrypted) == 0 {
			t.Error("Encrypt() returned empty data")
			return
		}

		decrypted, err := Decrypt(encrypted)
		if err != nil {
			t.Errorf("Decrypt() error = %v", err)
			return
		}

		if string(decrypted) != string(testData) {
			t.Errorf("Decrypt() = %v, want %v", string(decrypted), string(testData))
		}
	})

	t.Run("EncryptToBase64_DecryptFromBase64", func(t *testing.T) {
		if _, err := LoadPublicKey(keys.publicKeyFile); err != nil {
			t.Fatalf("failed to load public key: %v", err)
		}
		if _, err := LoadPrivateKey(keys.privateKeyFile); err != nil {
			t.Fatalf("failed to load private key: %v", err)
		}

		testData := []byte("test message for base64 encoding")
		encryptedBase64, err := EncryptToBase64(testData)
		if err != nil {
			t.Errorf("EncryptToBase64() error = %v", err)
			return
		}

		if encryptedBase64 == "" {
			t.Error("EncryptToBase64() returned empty string")
			return
		}

		decrypted, err := DecryptFromBase64(encryptedBase64)
		if err != nil {
			t.Errorf("DecryptFromBase64() error = %v", err)
			return
		}

		if string(decrypted) != string(testData) {
			t.Errorf("DecryptFromBase64() = %v, want %v", string(decrypted), string(testData))
		}
	})

	t.Run("DecryptWithoutPrivateKey", func(t *testing.T) {
		testData := []byte("encrypted data")
		_, err := Decrypt(testData)
		if err == nil {
			t.Error("Decrypt() expected error without private key, got nil")
		}
	})
}

func TestCryptoWithoutKeys(t *testing.T) {
	publicKey = nil
	privateKey = nil

	t.Run("EncryptWithoutPublicKey", func(t *testing.T) {
		publicKey = nil
		testData := []byte("test data")
		_, err := Encrypt(testData)
		if err == nil {
			t.Error("Encrypt() expected error without public key, got nil")
		}
	})

	t.Run("DecryptWithoutPrivateKey", func(t *testing.T) {
		privateKey = nil
		testData := []byte("encrypted data")
		_, err := Decrypt(testData)
		if err == nil {
			t.Error("Decrypt() expected error without private key, got nil")
		}
	})
}
