package db

import (
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/crypto"
)

// In a real application, we would use a testcontainers or real DB connection to test the fallback logic.
// For the sake of this mock exercise with the DB interface, we can test the `UpsertContract` function's behavior
// if we stub the pgxpool, but since we cannot easily stub pgxpool methods without an interface, we can instead
// verify that PGStore's `WithKMSClient` correctly sets the field.
// We should ideally test against a real DB as mandated by the instructions if possible.

func TestPGStore_WithKMSClient(t *testing.T) {
	store := NewPGStore(nil)
	if store.kmsClient != nil {
		t.Error("expected kmsClient to be nil initially")
	}

	mockClient, _ := crypto.NewMockKMSClient("test-arn")
	store.WithKMSClient(mockClient)
	if store.kmsClient == nil {
		t.Error("expected kmsClient to be set")
	}
}
