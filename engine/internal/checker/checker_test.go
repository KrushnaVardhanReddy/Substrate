// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package checker

import (
	"context"
	"errors"
	"testing"
)

func TestParseSchema(t *testing.T) {
	err := ParseSchema(context.Background(), func() error { return nil })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedErr := errors.New("parse error")
	err = ParseSchema(context.Background(), func() error { return expectedErr })
	if err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestCalculateDiff(t *testing.T) {
	err := CalculateDiff(context.Background(), func() error { return nil })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedErr := errors.New("diff error")
	err = CalculateDiff(context.Background(), func() error { return expectedErr })
	if err != expectedErr {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}
