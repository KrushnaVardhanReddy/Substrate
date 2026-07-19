package handlers

import (
	"encoding/json"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
	"net/http"
)

import (
	"context"
	"fmt"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/egress"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
)

type SaveDiffRequest struct {
	DiffReport        json.RawMessage `json:"diff_report"`
	IsAuditMode       bool            `json:"is_audit_mode"`
	Org               string          `json:"org"`
	ProviderRepo      string          `json:"provider_repo"`
	PRNumber          int             `json:"pr_number"`
	CommitSHA         string          `json:"commit_sha"`
	HeadSchemaContent string          `json:"head_schema_content"`
	SchemaType        string          `json:"schema_type"`
	ConfigContent     string          `json:"config_content"`
	InstallationID    int64           `json:"installation_id"`
}

type SaveDiffResponse struct {
	ID string `json:"id"`
}

func SaveDiffHandler(store db.Store, riverClient workers.JobEnqueuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SaveDiffRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if len(req.DiffReport) == 0 {
			http.Error(w, "diff_report is required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		id, err := store.SaveDiffReport(ctx, req.DiffReport, req.IsAuditMode, req.Org, req.ProviderRepo)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Check for breaking changes and dispatch webhook
		var diffReport services.DiffReport
		if err := json.Unmarshal(req.DiffReport, &diffReport); err == nil {
			if diffReport.Summary.BreakingCount > 0 && req.Org != "" {
				// We need to trigger webhook asynchronously
				go func(diffId string, diffReq SaveDiffRequest) {
					// Use a new context with timeout for background task
					bgCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
					defer cancel()

					var brokenConsumers []egress.BrokenConsumer
					// Perform cross-repo check if provider info is present
					if diffReq.ProviderRepo != "" {
						crReq := services.CrossRepoCheckRequest{
							InstallationID:    diffReq.InstallationID,
							Org:               diffReq.Org,
							ProviderRepo:      diffReq.ProviderRepo,
							HeadSchemaContent: diffReq.HeadSchemaContent,
							SchemaType:        diffReq.SchemaType,
							ConfigContent:     diffReq.ConfigContent,
						}

						crResp, err := services.PerformCrossRepoCheck(bgCtx, store, crReq)
						if err == nil {
							for _, res := range crResp.Results {
								if res.Status == "breaking" {
									brokenConsumers = append(brokenConsumers, egress.BrokenConsumer{
										Repo:       res.ConsumerRepo,
										Codeowners: []string{}, // Add codeowners if we had a way to resolve it
									})
								}
							}
						}
					}

					eventID := fmt.Sprintf("evt_%s_%d", diffId, time.Now().UnixNano())
					event := egress.BreakingChangeEvent{
						EventID:   eventID,
						EventType: "substrate.breaking_change.detected",
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						Data: egress.EventData{
							Organization:    diffReq.Org,
							ProviderRepo:    diffReq.ProviderRepo,
							PRNumber:        diffReq.PRNumber,
							CommitSHA:       diffReq.CommitSHA,
							BreakingCount:   diffReport.Summary.BreakingCount,
							BrokenConsumers: brokenConsumers,
							DiffURL:         fmt.Sprintf("https://substrate.%s/diff/%s", diffReq.Org, diffId), // Mock url for now
						},
					}

					egress.DispatchEvent(bgCtx, store, riverClient, event)

				}(id.String(), req)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(SaveDiffResponse{ID: id.String()})
	}
}

func GetDiffHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "invalid id format", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		diffReport, err := store.GetDiffReport(ctx, id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(diffReport)
	}
}
