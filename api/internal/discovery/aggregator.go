package discovery

import "strings"

type Edge struct {
	SourceRepo string
	TargetRepo string
	Confidence int
	Signal     string
}

func AggregateSignals(edges []Edge) []Edge {
	var finalEdges []Edge

	// Map to track repos by topic
	topics := make(map[string][]string) // topic -> list of repos

	for _, edge := range edges {
		if strings.HasPrefix(edge.Signal, "kafka_topic:") {
			topic := strings.TrimPrefix(edge.Signal, "kafka_topic:")
			topics[topic] = append(topics[topic], edge.SourceRepo)
		} else {
			finalEdges = append(finalEdges, edge)
		}
	}

	// Create edges between repos sharing the same topic
	for _, repos := range topics {
		if len(repos) > 1 {
			for i := 0; i < len(repos); i++ {
				for j := 0; j < len(repos); j++ {
					if i != j {
						finalEdges = append(finalEdges, Edge{
							SourceRepo: repos[i],
							TargetRepo: repos[j],
							Confidence: 20,
							Signal:     "kafka_topic_shared",
						})
					}
				}
			}
		} else if len(repos) == 1 {
            // Include it as a self-reference if it's the only one
			finalEdges = append(finalEdges, Edge{
				SourceRepo: repos[0],
				TargetRepo: repos[0],
				Confidence: 20,
				Signal:     "kafka_topic_shared",
			})
        }
	}

	return finalEdges
}
