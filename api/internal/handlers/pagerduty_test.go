package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
)

type mockStoreWithOrgIntegration struct {
	edges []db.DependencyEdge
}

func (m *mockStoreWithOrgIntegration) UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error { return nil }
func (m *mockStoreWithOrgIntegration) GetDependencyGraph(ctx context.Context, orgName string) ([]db.DependencyEdge, error) { return m.edges, nil }
func (m *mockStoreWithOrgIntegration) GetConsumersByProviderContract(ctx context.Context, providerContractID uuid.UUID) ([]db.ConsumerDependency, error) { return nil, nil }
func (m *mockStoreWithOrgIntegration) UpdateDependencyConfidence(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error { return nil }
func (m *mockStoreWithOrgIntegration) UpdateDependencyStatus(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, status string) error { return nil }
func (m *mockStoreWithOrgIntegration) CountDownstreamDependencies(ctx context.Context, providerFullName string) (int, error) { return 0, nil }

func (m *mockStoreWithOrgIntegration) GetOrgIDByName(ctx context.Context, orgName string) (uuid.UUID, error) { return uuid.New(), nil }
func (m *mockStoreWithOrgIntegration) UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error) { return uuid.New(), nil }
func (m *mockStoreWithOrgIntegration) CountReposByOrg(ctx context.Context, orgName string) (int, error) { return 0, nil }
func (m *mockStoreWithOrgIntegration) UpsertOrgIntegration(ctx context.Context, orgID uuid.UUID, provider, apiKeyEncrypted string) (db.OrgIntegration, error) { return db.OrgIntegration{}, nil }
func (m *mockStoreWithOrgIntegration) GetOrgIntegration(ctx context.Context, orgID uuid.UUID, provider string) (db.OrgIntegration, error) {
	return db.OrgIntegration{APIKeyEncrypted: "test-secret-key"}, nil
}

func (m *mockStoreWithOrgIntegration) UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error) { return uuid.New(), nil }
func (m *mockStoreWithOrgIntegration) ListReposByOrg(ctx context.Context, orgName string) ([]db.Repository, error) { return nil, nil }
func (m *mockStoreWithOrgIntegration) GetAllRepositories(ctx context.Context) ([]db.Repository, error) { return nil, nil }
func (m *mockStoreWithOrgIntegration) UpsertRepoMetric(ctx context.Context, repoID uuid.UUID, score int) error { return nil }
func (m *mockStoreWithOrgIntegration) GetTimeSinceLastBreak(ctx context.Context, repoID uuid.UUID) (*time.Time, error) { return nil, nil }
func (m *mockStoreWithOrgIntegration) UpsertRepoGuide(ctx context.Context, org, repo, filePath, title, content string) error { return nil }
func (m *mockStoreWithOrgIntegration) ListRepoGuides(ctx context.Context, org, repo string) ([]db.RepoGuide, error) { return nil, nil }
func (m *mockStoreWithOrgIntegration) GetRepoGuide(ctx context.Context, org, repo, slug string) (*db.RepoGuide, error) { return nil, nil }

func TestPagerDutyWebhookHandler(t *testing.T) {
	mockStore := &mockStoreWithOrgIntegration{
		edges: []db.DependencyEdge{
			{ProviderFullName: "test-org/service-a", ConsumerFullName: "test-org/service-b"},
			{ProviderFullName: "test-org/service-a", ConsumerFullName: "test-org/service-c"},
		},
	}

	handler := PagerDutyWebhookHandler(mockStore)

	t.Run("ValidPayload", func(t *testing.T) {
		payload := []byte(`{
			"event": {
				"data": {
					"id": "inc_123",
					"service": {
						"summary": "test-org/service-a"
					}
				}
			}
		}`)

		req := httptest.NewRequest("POST", "/api/v1/integrations/pagerduty/webhook", bytes.NewBuffer(payload))

		mac := hmac.New(sha256.New, []byte("test-secret-key"))
		mac.Write(payload)
		expectedMAC := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Webhook-Signature-256", "v1=" + expectedMAC)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&resp)

		impact := resp["downstream_impact"].([]interface{})
		if len(impact) != 2 {
			t.Errorf("expected 2 downstream services, got %v", len(impact))
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		payload := []byte(`{
			"event": {
				"data": {
					"id": "inc_123",
					"service": {
						"summary": "test-org/service-a"
					}
				}
			}
		}`)

		req := httptest.NewRequest("POST", "/api/v1/integrations/pagerduty/webhook", bytes.NewBuffer(payload))
		req.Header.Set("X-Webhook-Signature-256", "v1=invalid_signature")

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

    t.Run("MissingSignature", func(t *testing.T) {
		payload := []byte(`{
			"event": {
				"data": {
					"id": "inc_123",
					"service": {
						"summary": "test-org/service-a"
					}
				}
			}
		}`)

		req := httptest.NewRequest("POST", "/api/v1/integrations/pagerduty/webhook", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("InvalidPayload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/integrations/pagerduty/webhook", bytes.NewBuffer([]byte(`{`)))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})
}
