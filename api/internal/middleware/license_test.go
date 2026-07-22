package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"github.com/KrushnaVardhanReddy/substrate/engine/licensing"
)

func generateTestKeys() (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return privateKey, &privateKey.PublicKey
}

func createTestLicense(privateKey *rsa.PrivateKey, orgName string, features []string, expiresAt time.Time) string {
	claims := licensing.LicenseClaims{
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

func TestLicenseMiddleware(t *testing.T) {
	privateKey, publicKey := generateTestKeys()

	tests := []struct {
		name           string
		setupLicense   func() (string, func()) // returns path and cleanup
		expectedFeat   string
		expectedStatus bool // true if feature is expected to be present
	}{
		{
			name: "Valid license with features",
			setupLicense: func() (string, func()) {
				token := createTestLicense(privateKey, "Test Org", []string{"premium_feature"}, time.Now().Add(24*time.Hour))
				f, _ := os.CreateTemp("", "*.lic")
				f.WriteString(token)
				f.Close()
				return f.Name(), func() { os.Remove(f.Name()) }
			},
			expectedFeat:   "premium_feature",
			expectedStatus: true,
		},
		{
			name: "Valid license without requested feature",
			setupLicense: func() (string, func()) {
				token := createTestLicense(privateKey, "Test Org", []string{"other_feature"}, time.Now().Add(24*time.Hour))
				f, _ := os.CreateTemp("", "*.lic")
				f.WriteString(token)
				f.Close()
				return f.Name(), func() { os.Remove(f.Name()) }
			},
			expectedFeat:   "premium_feature",
			expectedStatus: false,
		},
		{
			name: "Expired license gracefully degrades",
			setupLicense: func() (string, func()) {
				token := createTestLicense(privateKey, "Test Org", []string{"premium_feature"}, time.Now().Add(-24*time.Hour))
				f, _ := os.CreateTemp("", "*.lic")
				f.WriteString(token)
				f.Close()
				return f.Name(), func() { os.Remove(f.Name()) }
			},
			expectedFeat:   "premium_feature",
			expectedStatus: false,
		},
		{
			name: "No license file gracefully degrades",
			setupLicense: func() (string, func()) {
				return "", func() {}
			},
			expectedFeat:   "premium_feature",
			expectedStatus: false,
		},
		{
			name: "Invalid license file gracefully degrades",
			setupLicense: func() (string, func()) {
				f, _ := os.CreateTemp("", "*.lic")
				f.WriteString("invalid.token.data")
				f.Close()
				return f.Name(), func() { os.Remove(f.Name()) }
			},
			expectedFeat:   "premium_feature",
			expectedStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setupLicense()
			defer cleanup()

			viper.Set("SUBSTRATE_LICENSE_FILE", path)
			defer viper.Reset()

			var actualStatus bool
			handler := LicenseMiddleware(publicKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualStatus = CheckFeature(r.Context(), tt.expectedFeat)
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if actualStatus != tt.expectedStatus {
				t.Errorf("CheckFeature(%s) = %v, want %v", tt.expectedFeat, actualStatus, tt.expectedStatus)
			}
		})
	}
}
