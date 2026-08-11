// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package licensing

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type LicenseClaims struct {
	OrgName  string   `json:"org_name"`
	Features []string `json:"features"`
	jwt.RegisteredClaims
}

type Validator struct {
	publicKey *rsa.PublicKey
}

func NewValidator(publicKey *rsa.PublicKey) *Validator {
	return &Validator{publicKey: publicKey}
}

func (v *Validator) ValidateLicenseFile(path string) (*LicenseClaims, error) {
	tokenString, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file: %w", err)
	}
	return v.ValidateLicense(strings.TrimSpace(string(tokenString)))
}

func (v *Validator) ValidateLicense(tokenString string) (*LicenseClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &LicenseClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*LicenseClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token or claims")
}
