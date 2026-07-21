package compliance

import (
	"encoding/json"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/stretchr/testify/assert"
)

func TestAuditPIIDetection(t *testing.T) {
	tests := []struct {
		name          string
		changes       []report.Change
		expectedAlert int
		expectedTypes []string
	}{
		{
			name: "Detect SSN",
			changes: []report.Change{
				{Path: "/users/ssn"},
			},
			expectedAlert: 1,
			expectedTypes: []string{"PII:SSN"},
		},
		{
			name: "Detect Password",
			changes: []report.Change{
				{Path: "password_hash"},
			},
			expectedAlert: 1,
			expectedTypes: []string{"SECURITY:CREDENTIAL"},
		},
		{
			name: "Detect Multiple",
			changes: []report.Change{
				{Path: "/patient/medical_history"},
				{Path: "email"},
			},
			expectedAlert: 2,
			expectedTypes: []string{"HIPAA:PHI", "PII:CONTACT"},
		},
		{
			name: "No Match",
			changes: []report.Change{
				{Path: "id"},
				{Path: "created_at"},
			},
			expectedAlert: 0,
			expectedTypes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := &report.DiffReport{
				BreakingChanges: tt.changes,
			}
			Audit(rep)

			assert.Len(t, rep.ComplianceAlerts, tt.expectedAlert)

			for i, alert := range rep.ComplianceAlerts {
				assert.Equal(t, tt.expectedTypes[i], alert.ComplianceType)
			}
		})
	}
}

func TestNotifySecurityTeam(t *testing.T) {
	var receivedPayload map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	os.Setenv("SUBSTRATE_SLACK_WEBHOOK", server.URL)
	defer os.Unsetenv("SUBSTRATE_SLACK_WEBHOOK")

	alert := report.ComplianceAlert{
		Path:           "/user/ssn",
		ComplianceType: "PII:SSN",
		Message:        "Detected field matching PII:SSN pattern",
	}

	notifySecurityTeam(alert)

	assert.NotNil(t, receivedPayload)
	text, ok := receivedPayload["text"].(string)
	assert.True(t, ok)
	assert.Contains(t, text, "PII:SSN")
	assert.Contains(t, text, "/user/ssn")
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
