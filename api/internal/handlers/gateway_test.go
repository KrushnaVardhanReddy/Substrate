package handlers

import (
	"testing"
)

func TestGatewaySyncHandler(t *testing.T) {
	handler := GatewaySyncHandler(nil, nil)
	if handler == nil {
		t.Fatal("expected handler to not be nil")
	}
}
