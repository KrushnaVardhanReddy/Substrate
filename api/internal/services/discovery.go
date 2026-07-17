package services

import (
	"strings"
)

// DiscoverImplicitDependencies parses manifest files to identify infrastructure dependencies heuristically.
func DiscoverImplicitDependencies(files map[string]string) []DependencyPayload {
	var deps []DependencyPayload

	// Use a map to track discovered dependencies to avoid duplicates
	// key is providerRepo name, e.g. "infra:postgres"
	discovered := make(map[string]bool)

	addDep := func(provider string) {
		if !discovered[provider] {
			deps = append(deps, DependencyPayload{
				ProviderRepo:         provider,
				ProviderGithubRepoID: 0, // Deterministic mock ID for implicit deps
				SchemaType:           "infrastructure",
				SpecPath:             "implicit",
				Branch:               "main",
				RawContent:           "{}",
			})
			discovered[provider] = true
		}
	}

	for fileName, content := range files {
		contentLower := strings.ToLower(content)

		switch fileName {
		case "package.json":
			if strings.Contains(contentLower, "\"pg\"") || strings.Contains(contentLower, "\"pq\"") || strings.Contains(contentLower, "\"psycopg2\"") {
				addDep("infra:postgres")
			}
			if strings.Contains(contentLower, "\"mongoose\"") || strings.Contains(contentLower, "\"mongodb\"") {
				addDep("infra:mongodb")
			}
			if strings.Contains(contentLower, "\"ioredis\"") || strings.Contains(contentLower, "\"redis\"") {
				addDep("infra:redis")
			}
			if strings.Contains(contentLower, "\"kafkajs\"") {
				addDep("infra:kafka")
			}
			if strings.Contains(contentLower, "\"amqplib\"") {
				addDep("infra:rabbitmq")
			}
			if strings.Contains(contentLower, "\"stripe-node\"") {
				addDep("saas:stripe")
			}
			if strings.Contains(contentLower, "\"twilio\"") {
				addDep("saas:twilio")
			}
			if strings.Contains(contentLower, "\"@slack/web-api\"") {
				addDep("saas:slack")
			}
			if strings.Contains(contentLower, "\"@aws-sdk/client-s3\"") {
				addDep("saas:aws-s3")
			}
		case "go.mod":
			if strings.Contains(contentLower, "github.com/lib/pq") || strings.Contains(contentLower, "github.com/jackc/pgx") {
				addDep("infra:postgres")
			}
			if strings.Contains(contentLower, "go.mongodb.org/mongo-driver") {
				addDep("infra:mongodb")
			}
			if strings.Contains(contentLower, "github.com/go-redis/redis") || strings.Contains(contentLower, "github.com/redis/go-redis") {
				addDep("infra:redis")
			}
			if strings.Contains(contentLower, "github.com/confluentinc/confluent-kafka-go") {
				addDep("infra:kafka")
			}
			if strings.Contains(contentLower, "github.com/streadway/amqp") || strings.Contains(contentLower, "github.com/rabbitmq/amqp091-go") {
				addDep("infra:rabbitmq")
			}
			if strings.Contains(contentLower, "github.com/stripe/stripe-go") {
				addDep("saas:stripe")
			}
			if strings.Contains(contentLower, "github.com/twilio/twilio-go") {
				addDep("saas:twilio")
			}
			if strings.Contains(contentLower, "github.com/slack-go/slack") {
				addDep("saas:slack")
			}
			if strings.Contains(contentLower, "github.com/aws/aws-sdk-go-v2/service/s3") {
				addDep("saas:aws-s3")
			}
		case "docker-compose.yml":
			if strings.Contains(contentLower, "image: postgres") || strings.Contains(contentLower, "image: bitnami/postgresql") {
				addDep("infra:postgres")
			}
			if strings.Contains(contentLower, "image: mongo") || strings.Contains(contentLower, "image: bitnami/mongodb") {
				addDep("infra:mongodb")
			}
			if strings.Contains(contentLower, "image: redis") || strings.Contains(contentLower, "image: bitnami/redis") {
				addDep("infra:redis")
			}
			if strings.Contains(contentLower, "image: confluentinc/cp-kafka") || strings.Contains(contentLower, "image: bitnami/kafka") {
				addDep("infra:kafka")
			}
			if strings.Contains(contentLower, "image: rabbitmq") || strings.Contains(contentLower, "image: bitnami/rabbitmq") {
				addDep("infra:rabbitmq")
			}
		}
	}

	return deps
}
