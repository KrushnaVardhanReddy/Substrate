package egress

import (
	"context"
	"log"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
)

// We alias types to services package to avoid backwards compatibility breaks for other packages importing egress types.
type EventData = services.EventData
type BrokenConsumer = services.BrokenConsumer
type BreakingChangeEvent = services.BreakingChangeEvent

func DispatchEvent(ctx context.Context, store db.Store, riverClient workers.JobEnqueuer, event services.BreakingChangeEvent) error {
	webhooks, err := store.GetWebhooks(ctx, event.Data.Organization)
	if err != nil {
		log.Printf("Error fetching webhooks for org %s: %v", event.Data.Organization, err)
		return err
	}

	if len(webhooks) == 0 {
		return nil
	}

	// For multiple webhooks, it's safer to just insert them individually.
	// Since we are offloading to River, we can create multiple EgressWebhookJob.
	for _, hook := range webhooks {
		_, err := riverClient.Insert(ctx, workers.EgressWebhookJob{
			Hook:  hook,
			Event: event,
		}, nil)
		if err != nil {
			log.Printf("Failed to enqueue EgressWebhookJob for webhook %s: %v", hook.URL, err)
		}
	}

	return nil
}
