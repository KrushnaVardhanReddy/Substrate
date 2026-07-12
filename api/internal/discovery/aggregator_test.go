package discovery

import (
	"reflect"
	"testing"
)

func TestAggregateSignals(t *testing.T) {
	tests := []struct {
		name  string
		edges []Edge
		want  []Edge
	}{
		{
			name: "map producers and consumers via shared Topic names",
			edges: []Edge{
				{
					SourceRepo: "myorg/orders-service",
					TargetRepo: "myorg/orders-service",
					Confidence: 20,
					Signal:     "kafka_topic:orders.v2.order_created",
				},
				{
					SourceRepo: "myorg/payments-service",
					TargetRepo: "myorg/payments-service",
					Confidence: 20,
					Signal:     "kafka_topic:orders.v2.order_created",
				},
			},
			want: []Edge{
				{
					SourceRepo: "myorg/orders-service",
					TargetRepo: "myorg/payments-service",
					Confidence: 20,
					Signal:     "kafka_topic_shared",
				},
				{
					SourceRepo: "myorg/payments-service",
					TargetRepo: "myorg/orders-service",
					Confidence: 20,
					Signal:     "kafka_topic_shared",
				},
			},
		},
		{
			name: "passthrough other edges",
			edges: []Edge{
				{
					SourceRepo: "myorg/a",
					TargetRepo: "myorg/b",
					Confidence: 35,
					Signal:     "asyncapi_$ref",
				},
			},
			want: []Edge{
				{
					SourceRepo: "myorg/a",
					TargetRepo: "myorg/b",
					Confidence: 35,
					Signal:     "asyncapi_$ref",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateSignals(tt.edges)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AggregateSignals() = %v, want %v", got, tt.want)
			}
		})
	}
}
