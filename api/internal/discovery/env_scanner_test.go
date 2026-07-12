package discovery

import (
	"reflect"
	"testing"
)

func TestScanEnvFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []DiscoveredDependency
	}{
		{
			name: "high signal urls",
			content: `
USERS_API_URL=https://users.myorg.com
PAYMENTS_ENDPOINT="https://api.payments.myorg.com/v2"
DATABASE_URL=postgres://localhost
# comment here
FRONTEND_BASE_URL=http://localhost:3000
SECRET_API_URL=https://secret.com
`,
			want: []DiscoveredDependency{
				{VarName: "USERS_API_URL", VarValue: "https://users.myorg.com", ConfidenceScore: 15},
				{VarName: "PAYMENTS_ENDPOINT", VarValue: "https://api.payments.myorg.com/v2", ConfidenceScore: 15},
				{VarName: "FRONTEND_BASE_URL", VarValue: "http://localhost:3000", ConfidenceScore: 15},
			},
		},
		{
			name: "secrets excluded",
			content: `
USERS_API_KEY=1234
PAYMENTS_SECRET=secret
MY_API_URL=ok
`,
			want: []DiscoveredDependency{
				{VarName: "MY_API_URL", VarValue: "ok", ConfidenceScore: 15},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScanEnvFile(tt.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanEnvFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanDockerCompose(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []DiscoveredDependency
	}{
		{
			name: "docker compose list and dict",
			content: `
services:
  frontend:
    environment:
      - USERS_API_URL=http://users-service:8080
      - SECRET_KEY=abc
    ports:
      - "3000:3000"
  backend:
    environment:
      PAYMENTS_ENDPOINT: "https://api.payments.myorg.com"
      DB_URL: postgres://db
`,
			want: []DiscoveredDependency{
				{VarName: "USERS_API_URL", VarValue: "http://users-service:8080", ConfidenceScore: 15},
				{VarName: "PAYMENTS_ENDPOINT", VarValue: "https://api.payments.myorg.com", ConfidenceScore: 15},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScanDockerCompose(tt.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanDockerCompose() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanKubernetesManifest(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []DiscoveredDependency
	}{
		{
			name: "kubernetes env vars",
			content: `
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: myapp
          env:
            - name: USERS_API_URL
              value: "http://users.default.svc.cluster.local"
            - name: PAYMENTS_ENDPOINT
              value: https://api.payments.com
            - name: MY_SECRET
              value: "secret"
`,
			want: []DiscoveredDependency{
				{VarName: "USERS_API_URL", VarValue: "http://users.default.svc.cluster.local", ConfidenceScore: 15},
				{VarName: "PAYMENTS_ENDPOINT", VarValue: "https://api.payments.com", ConfidenceScore: 15},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScanKubernetesManifest(tt.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanKubernetesManifest() = %v, want %v", got, tt.want)
			}
		})
	}
}
