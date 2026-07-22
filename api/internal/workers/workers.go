package workers

import (
	"context"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/discovery"
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
	Store       db.Store
	GHClient    github.Client
	RiverClient JobEnqueuer
}

func (w *PushWebhookWorker) Work(ctx context.Context, job *river.Job[PushWebhookJob]) error {
	_, err := services.ProcessPush(ctx, w.Store, w.GHClient, job.Args.Payload)
	if err == nil && w.RiverClient != nil {
		_, _ = w.RiverClient.Insert(ctx, DiscoveryJob{
			Payload: job.Args.Payload,
		}, nil)
	}
	return err
}

type DiscoveryJob struct {
	Payload services.PushPayload `json:"payload"`
}

func (DiscoveryJob) Kind() string { return "discovery_job" }

type DiscoveryWorker struct {
	river.WorkerDefaults[DiscoveryJob]
	Store    db.Store
	GHClient github.Client
}

func (w *DiscoveryWorker) Work(ctx context.Context, job *river.Job[DiscoveryJob]) error {
	payload := job.Args.Payload

	parts := strings.Split(payload.Repo, "/")
	if len(parts) != 2 {
		return nil
	}
	owner, repo := parts[0], parts[1]

	scanner := discovery.NewRepositoryScanner(w.GHClient)
	urlResolver := func(url string) string { return url }
	edges, err := scanner.ScanRepository(ctx, owner, repo, urlResolver)
	if err != nil {
		return err
	}

	return services.ProcessDiscoveredEdges(ctx, w.Store, payload, edges)
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

func RegisterWorkers(store db.Store, ghClient github.Client) (*river.Workers, *PushWebhookWorker) {
	workers := river.NewWorkers()
	pushWorker := &PushWebhookWorker{Store: store, GHClient: ghClient}
	river.AddWorker(workers, &SyncWebhookWorker{Store: store})
	river.AddWorker(workers, pushWorker)
	river.AddWorker(workers, &CrossRepoCheckWorker{Store: store})
	river.AddWorker(workers, &EgressWebhookWorker{})
	river.AddWorker(workers, &DiscoveryWorker{Store: store, GHClient: ghClient})
	river.AddWorker(workers, &RecalculateRiskScoresWorker{store: store})
	return workers, pushWorker
}
