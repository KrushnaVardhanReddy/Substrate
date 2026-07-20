package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

var (
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrEncryptionFailed = errors.New("encryption failed")
)

type KMSClient interface {
	Encrypt(plaintext []byte) (ciphertext []byte, keyARN string, err error)
	Decrypt(ciphertext []byte, keyARN string) (plaintext []byte, err error)
}

type MockKMSClient struct {
	MasterKey []byte
	KeyARN    string
}

func NewMockKMSClient(keyARN string) (*MockKMSClient, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return &MockKMSClient{
		MasterKey: key,
		KeyARN:    keyARN,
	}, nil
}

func (m *MockKMSClient) Encrypt(plaintext []byte) ([]byte, string, error) {
	block, err := aes.NewCipher(m.MasterKey)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, m.KeyARN, nil
}

func (m *MockKMSClient) Decrypt(ciphertext []byte, keyARN string) ([]byte, error) {
	if keyARN != m.KeyARN {
		return nil, fmt.Errorf("invalid key ARN: expected %s, got %s", m.KeyARN, keyARN)
	}

	block, err := aes.NewCipher(m.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("%w: ciphertext too short", ErrDecryptionFailed)
	}

	nonce, ciphertextPart := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextPart, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

// Ensure MockKMSClient implements KMSClient
var _ KMSClient = (*MockKMSClient)(nil)
