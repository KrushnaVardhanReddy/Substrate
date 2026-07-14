package workers

import (
	"context"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/riverqueue/river"
)

type SyncWebhookJob struct {
	Request services.SyncRequest `json:"request"`
}

func (SyncWebhookJob) Kind() string { return "sync_webhook" }

type SyncWebhookWorker struct {
	river.WorkerDefaults[SyncWebhookJob]
	Store db.Store
}

func (w *SyncWebhookWorker) Work(ctx context.Context, job *river.Job[SyncWebhookJob]) error {
	_, err := services.ProcessSync(ctx, w.Store, job.Args.Request)
	return err
}

type PushWebhookJob struct {
	Payload services.PushPayload `json:"payload"`
}

func (PushWebhookJob) Kind() string { return "push_webhook" }

type PushWebhookWorker struct {
	river.WorkerDefaults[PushWebhookJob]
	Store    db.Store
	GHClient github.Client
}

func (w *PushWebhookWorker) Work(ctx context.Context, job *river.Job[PushWebhookJob]) error {
	_, err := services.ProcessPush(ctx, w.Store, w.GHClient, job.Args.Payload)
	return err
}

type CrossRepoCheckJob struct {
	Request services.CrossRepoCheckRequest `json:"request"`
}

func (CrossRepoCheckJob) Kind() string { return "cross_repo_check" }

type CrossRepoCheckWorker struct {
	river.WorkerDefaults[CrossRepoCheckJob]
	Store db.Store
}

func (w *CrossRepoCheckWorker) Work(ctx context.Context, job *river.Job[CrossRepoCheckJob]) error {
	_, err := services.PerformCrossRepoCheck(ctx, w.Store, job.Args.Request)
	return err
}

type EgressWebhookJob struct {
	Hook  db.WebhookConfig             `json:"hook"`
	Event services.BreakingChangeEvent `json:"event"`
}

func (EgressWebhookJob) Kind() string { return "egress_webhook" }

type EgressWebhookWorker struct {
	river.WorkerDefaults[EgressWebhookJob]
}

func (w *EgressWebhookWorker) Work(ctx context.Context, job *river.Job[EgressWebhookJob]) error {
	return services.DispatchWebhook(ctx, job.Args.Hook, job.Args.Event)
}

func RegisterWorkers(store db.Store, ghClient github.Client) *river.Workers {
	workers := river.NewWorkers()
	river.AddWorker(workers, &SyncWebhookWorker{Store: store})
	river.AddWorker(workers, &PushWebhookWorker{Store: store, GHClient: ghClient})
	river.AddWorker(workers, &CrossRepoCheckWorker{Store: store})
	river.AddWorker(workers, &EgressWebhookWorker{})
	return workers
}
