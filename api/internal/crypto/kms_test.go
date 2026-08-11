// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package crypto

import (
	"bytes"
	"testing"
)

func TestMockKMSClient_RoundTrip(t *testing.T) {
	keyARN := "arn:aws:kms:us-east-1:123456789012:key/mrk-12345"
	client, err := NewMockKMSClient(keyARN)
	if err != nil {
		t.Fatalf("failed to create mock KMS client: %v", err)
	}

	plaintext := []byte("this is a highly sensitive schema payload")

	ciphertext, arn, err := client.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	if arn != keyARN {
		t.Errorf("expected ARN %s, got %s", keyARN, arn)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext is identical to plaintext")
	}

	decrypted, err := client.Decrypt(ciphertext, arn)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("expected decrypted text %q, got %q", string(plaintext), string(decrypted))
	}
}

func TestMockKMSClient_InvalidARN(t *testing.T) {
	keyARN := "arn:aws:kms:us-east-1:123456789012:key/mrk-12345"
	client, err := NewMockKMSClient(keyARN)
	if err != nil {
		t.Fatalf("failed to create mock KMS client: %v", err)
	}

	plaintext := []byte("this is a highly sensitive schema payload")
	ciphertext, _, err := client.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	_, err = client.Decrypt(ciphertext, "arn:aws:kms:us-east-1:123456789012:key/mrk-wrong")
	if err == nil {
		t.Error("expected error when decrypting with wrong ARN, got nil")
	}
}
