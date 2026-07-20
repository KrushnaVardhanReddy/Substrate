package deprecation

import (
	"testing"
)

func TestDeprecationConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      DeprecationConfig
		expected string
		expDate  string
	}{
		{
			name: "valid endpoint and date",
			cfg: DeprecationConfig{
				Endpoint:   "/api/v1/users",
				SunsetDate: "2024-12-31",
			},
			expected: "/api/v1/users",
			expDate:  "2024-12-31",
		},
		{
			name: "empty config",
			cfg: DeprecationConfig{
				Endpoint:   "",
				SunsetDate: "",
			},
			expected: "",
			expDate:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfg.Endpoint != tt.expected {
				t.Errorf("Expected endpoint %s, got %s", tt.expected, tt.cfg.Endpoint)
			}
			if tt.cfg.SunsetDate != tt.expDate {
				t.Errorf("Expected sunset_date %s, got %s", tt.expDate, tt.cfg.SunsetDate)
			}
		})
	}
}
