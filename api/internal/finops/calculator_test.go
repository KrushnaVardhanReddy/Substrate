package finops

import (
	"math"
	"testing"
)

func TestEstimateResponseByteSize(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]interface{}
		expected int64
	}{
		{
			name:     "nil schema",
			schema:   nil,
			expected: 0,
		},
		{
			name: "string type",
			schema: map[string]interface{}{
				"type": "string",
			},
			expected: 50,
		},
		{
			name: "integer type",
			schema: map[string]interface{}{
				"type": "integer",
			},
			expected: 8,
		},
		{
			name: "number type",
			schema: map[string]interface{}{
				"type": "number",
			},
			expected: 8,
		},
		{
			name: "boolean type",
			schema: map[string]interface{}{
				"type": "boolean",
			},
			expected: 1,
		},
		{
			name: "array of strings",
			schema: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			expected: 500, // 10 * 50
		},
		{
			name: "simple object",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "integer"},
					"name": map[string]interface{}{"type": "string"},
					"isActive": map[string]interface{}{"type": "boolean"},
				},
			},
			expected: 8 + 50 + 1,
		},
		{
			name: "nested object without type string",
			schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"nested": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"field1": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateResponseByteSize(tt.schema)
			if got != tt.expected {
				t.Errorf("EstimateResponseByteSize() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCalculateMonthlyCostDiff(t *testing.T) {
	tests := []struct {
		name          string
		baseBytes     int64
		proposedBytes int64
		rps           float64
		costPerGB     float64
		expected      float64
	}{
		{
			name:          "no difference",
			baseBytes:     100,
			proposedBytes: 100,
			rps:           100.0,
			costPerGB:     0.09,
			expected:      0.0,
		},
		{
			name:          "increase in size",
			baseBytes:     100,
			proposedBytes: 150, // diff = 50 bytes
			rps:           100.0,
			costPerGB:     0.09,
			// 50 * 100 * 3600 * 730 = 13,140,000,000 bytes
			// 13,140,000,000 / (1024^3) = ~12.237 GB
			// 12.237 * 0.09 = ~1.101
			expected: 1.101, // approximate, we will check with epsilon
		},
		{
			name:          "decrease in size",
			baseBytes:     150,
			proposedBytes: 100, // diff = -50 bytes
			rps:           100.0,
			costPerGB:     0.09,
			expected:      -1.101,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateMonthlyCostDiff(tt.baseBytes, tt.proposedBytes, tt.rps, tt.costPerGB)
			if math.Abs(got-tt.expected) > 0.01 { // Check within 1 cent tolerance
				t.Errorf("CalculateMonthlyCostDiff() = %v, want %v", got, tt.expected)
			}
		})
	}
}
