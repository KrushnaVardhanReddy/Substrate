package finops

// MetricsProvider defines an interface for fetching traffic metrics.
type MetricsProvider interface {
	GetEndpointRPS(path string) (float64, error)
}

// MockDatadogProvider is a mock implementation of MetricsProvider.
type MockDatadogProvider struct {
	// Credentials can be added here later
}

// NewMockDatadogProvider creates a new MockDatadogProvider.
func NewMockDatadogProvider() *MockDatadogProvider {
	return &MockDatadogProvider{}
}

// GetEndpointRPS returns a static RPS value if no real credentials are provided.
func (m *MockDatadogProvider) GetEndpointRPS(path string) (float64, error) {
	// Static RPS for now
	return 100.0, nil
}
