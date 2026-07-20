package db

import (
	"testing"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/crypto"
)

// In lieu of testing pgxpool directly without a test database setup,
// we test that the fallback logic in a store implementation preserves the contracts.
// Since MockStore implements Store, we can provide a small test verifying it works.
// However, the database fallback logic itself lives inside PGStore's queries.
// Due to missing real DB setup in tests (e.g. pgtest or testcontainers) for now,
// we verify the structure and client assignment.

func TestFallbackLogicMock(t *testing.T) {
	mockClient, _ := crypto.NewMockKMSClient("arn:aws:kms:us-east-1:111122223333:key/mrk-123")
	store := NewPGStore(nil).WithKMSClient(mockClient)

	if store.kmsClient == nil {
		t.Fatal("kmsClient not set")
	}

	// Test the client handles encryption to simulate what PGStore would do on Upsert
	plaintext := []byte("schema")
	ciphertext, arn, err := store.kmsClient.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	decrypted, err := store.kmsClient.Decrypt(ciphertext, arn)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if string(decrypted) != "schema" {
		t.Errorf("decrypted output mismatch: got %s", string(decrypted))
	}
}
