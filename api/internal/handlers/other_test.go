package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
)

func TestGraphHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/graph/myorg", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetPathValue("org", "myorg")

	rr := httptest.NewRecorder()
	handler := GraphHandler(&db.MockStore{})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestReposHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/repos/myorg", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetPathValue("org", "myorg")

	rr := httptest.NewRecorder()
	handler := ReposHandler(&db.MockStore{})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGraphHandler_MissingOrg(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/graph/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetPathValue("org", "")

	rr := httptest.NewRecorder()
	handler := GraphHandler(&db.MockStore{})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestReposHandler_MissingOrg(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/repos/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetPathValue("org", "")

	rr := httptest.NewRecorder()
	handler := ReposHandler(&db.MockStore{})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
