// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
)

func TestMCPProfileHandler_CreateProfile(t *testing.T) {
	mockStore := &db.MockStore{}
	handler := &MCPProfileHandler{Store: mockStore}

	reqBody := []byte(`{"org": "test-org", "name": "Test Profile", "allowed_tools": ["tool1"], "hitl_enabled": true}`)
	req := httptest.NewRequest("POST", "/api/v1/mcp/profiles", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateProfile(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}

func TestMCPProfileHandler_ListProfiles(t *testing.T) {
	mockStore := &db.MockStore{}
	handler := &MCPProfileHandler{Store: mockStore}

	req := httptest.NewRequest("GET", "/api/v1/mcp/profiles/test-org", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("org", "test-org")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ListProfiles(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestMCPProfileHandler_DeleteProfile(t *testing.T) {
	mockStore := &db.MockStore{}
	handler := &MCPProfileHandler{Store: mockStore}

	req := httptest.NewRequest("DELETE", "/api/v1/mcp/profiles/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.DeleteProfile(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestMCPProfileHandler_ListHITLQueue(t *testing.T) {
	mockStore := &db.MockStore{}
	handler := &MCPProfileHandler{Store: mockStore}

	req := httptest.NewRequest("GET", "/api/v1/mcp/hitl-queue/test-org", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("org", "test-org")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ListHITLQueue(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestMCPProfileHandler_ResolveHITLItem(t *testing.T) {
	mockStore := &db.MockStore{}
	handler := &MCPProfileHandler{Store: mockStore}

	reqBody := []byte(`{"status": "approved"}`)
	req := httptest.NewRequest("POST", "/api/v1/mcp/hitl-queue/1/resolve", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ResolveHITLItem(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
