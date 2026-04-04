package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type Ed25519Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewEd25519Signer(privateHex string) (*Ed25519Signer, error) {
	privBytes, err := hex.DecodeString(privateHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex private key: %w", err)
	}
	if len(privBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key length: got %d, want %d", len(privBytes), ed25519.PrivateKeySize)
	}
	
	// Извлекаем публичный ключ из приватного
	pubKey := privBytes[32:]
	return &Ed25519Signer{
		privateKey: ed25519.PrivateKey(privBytes),
		publicKey:  ed25519.PublicKey(pubKey),
	}, nil
}

func (s *Ed25519Signer) Sign(data []byte) (string, error) {
	sig := ed25519.Sign(s.privateKey, data)
	return base64.StdEncoding.EncodeToString(sig), nil
}

func VerifyEd25519Signature(pubHex string, data []byte, sigBase64 string) error {
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return fmt.Errorf("invalid public key hex: %w", err)
	}
	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return fmt.Errorf("invalid signature base64: %w", err)
	}
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), data, sigBytes) {
		return fmt.Errorf("signature verification failed: invalid signature")
	}
	return nil
}

// GetPublicKeyHex возвращает публичный ключ в hex для хранения в БД/выдачи вузу
func (s *Ed25519Signer) GetPublicKeyHex() string {
	return hex.EncodeToString(s.publicKey)
}