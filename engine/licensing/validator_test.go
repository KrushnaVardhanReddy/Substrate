// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package licensing

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestKeys() (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return privateKey, &privateKey.PublicKey
}

func createTestLicense(privateKey *rsa.PrivateKey, orgName string, features []string, expiresAt time.Time) string {
	claims := LicenseClaims{
		OrgName:  orgName,
		Features: features,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}
	return tokenString
}

func TestValidator_ValidateLicense(t *testing.T) {
	privateKey, publicKey := generateTestKeys()
	otherPrivateKey, _ := generateTestKeys()

	validator := NewValidator(publicKey)

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
		checkClaims func(*testing.T, *LicenseClaims)
	}{
		{
			name:        "Valid license",
			tokenString: createTestLicense(privateKey, "Test Org", []string{"feature_a", "feature_b"}, time.Now().Add(24*time.Hour)),
			wantErr:     false,
			checkClaims: func(t *testing.T, claims *LicenseClaims) {
				if claims.OrgName != "Test Org" {
					t.Errorf("expected OrgName 'Test Org', got '%s'", claims.OrgName)
				}
				if len(claims.Features) != 2 || claims.Features[0] != "feature_a" || claims.Features[1] != "feature_b" {
					t.Errorf("expected features [feature_a, feature_b], got %v", claims.Features)
				}
			},
		},
		{
			name:        "Expired license",
			tokenString: createTestLicense(privateKey, "Test Org", []string{"feature_a"}, time.Now().Add(-24*time.Hour)),
			wantErr:     true,
			checkClaims: nil,
		},
		{
			name:        "Wrong signature",
			tokenString: createTestLicense(otherPrivateKey, "Test Org", []string{"feature_a"}, time.Now().Add(24*time.Hour)),
			wantErr:     true,
			checkClaims: nil,
		},
		{
			name:        "Invalid token format",
			tokenString: "not.a.jwt",
			wantErr:     true,
			checkClaims: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := validator.ValidateLicense(tt.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLicense() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkClaims != nil {
				tt.checkClaims(t, claims)
			}
		})
	}
}

func TestValidator_ValidateLicenseFile(t *testing.T) {
	privateKey, publicKey := generateTestKeys()
	validator := NewValidator(publicKey)

	validToken := createTestLicense(privateKey, "File Org", []string{"enterprise"}, time.Now().Add(24*time.Hour))

	tmpFile, err := os.CreateTemp("", "license-*.lic")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(validToken); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	claims, err := validator.ValidateLicenseFile(tmpFile.Name())
	if err != nil {
		t.Errorf("ValidateLicenseFile() error = %v, wantErr false", err)
	}
	if claims == nil || claims.OrgName != "File Org" {
		t.Errorf("ValidateLicenseFile() invalid claims returned")
	}

	_, err = validator.ValidateLicenseFile("nonexistent.lic")
	if err == nil {
		t.Errorf("ValidateLicenseFile() expected error for nonexistent file, got nil")
	}
}
