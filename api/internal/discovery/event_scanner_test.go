// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package discovery

import (
	"reflect"
	"testing"
)

func TestEventScanner_ScanAsyncAPI(t *testing.T) {
	scanner := NewEventScanner()

	tests := []struct {
		name       string
		content    []byte
		sourceRepo string
		want       []Edge
	}{
		{
			name:       "valid asyncapi yaml with github ref",
			sourceRepo: "myorg/payments-service",
			content: []byte(`
channels:
  orders.v2.order_created:
    subscribe:
      message:
        $ref: 'https://github.com/myorg/orders-service/blob/main/schemas/order_created.avsc'
`),
			want: []Edge{
				{
					SourceRepo: "myorg/payments-service",
					TargetRepo: "myorg/orders-service",
					Confidence: 35,
					Signal:     "asyncapi_$ref",
				},
			},
		},
		{
			name:       "json with ref",
			sourceRepo: "myorg/payments-service",
			content: []byte(`{
				"channels": {
					"orders.v2.order_created": {
						"subscribe": {
							"message": {
								"$ref": "https://github.com/myorg/orders-service/blob/main/schemas/order_created.avsc"
							}
						}
					}
				}
			}`),
			want: []Edge{
				{
					SourceRepo: "myorg/payments-service",
					TargetRepo: "myorg/orders-service",
					Confidence: 35,
					Signal:     "asyncapi_$ref",
				},
			},
		},
		{
			name:       "ignore self refs",
			sourceRepo: "myorg/orders-service",
			content: []byte(`
channels:
  orders.v2.order_created:
    subscribe:
      message:
        $ref: 'https://github.com/myorg/orders-service/blob/main/schemas/order_created.avsc'
`),
			want: nil, // self refs are not cross-repo dependencies
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanner.ScanAsyncAPI(tt.content, tt.sourceRepo)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanAsyncAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventScanner_ScanTerraformKafka(t *testing.T) {
	scanner := NewEventScanner()

	tests := []struct {
		name       string
		content    []byte
		sourceRepo string
		want       []Edge
	}{
		{
			name:       "confluent kafka topic block",
			sourceRepo: "myorg/orders-service",
			content: []byte(`
resource "confluent_kafka_topic" "order_events" {
  kafka_cluster {
    id = "foo"
  }
  topic_name = "orders.v2.order_created"
}
`),
			want: []Edge{
				{
					SourceRepo: "myorg/orders-service",
					TargetRepo: "myorg/orders-service",
					Confidence: 20,
					Signal:     "kafka_topic:orders.v2.order_created",
				},
			},
		},
		{
			name:       "no match when no topic name",
			sourceRepo: "myorg/orders-service",
			content: []byte(`
resource "confluent_kafka_topic" "order_events" {
  partitions_count = 1
}
`),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanner.ScanTerraformKafka(tt.content, tt.sourceRepo)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanTerraformKafka() = %v, want %v", got, tt.want)
			}
		})
	}
}
