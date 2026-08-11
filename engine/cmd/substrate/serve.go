// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/graphql"
	"github.com/KrushnaVardhanReddy/substrate/engine/gates"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	sqlpkg "github.com/KrushnaVardhanReddy/substrate/engine/internal/sql"
)

type GateRequest struct {
	Gate          string `json:"gate"`
	BreakingCount int    `json:"breaking_count"`
	WarningCount  int    `json:"warning_count"`
	CrossRepoSafe bool   `json:"cross_repo_safe"`
}

type GateResponse struct {
	Pass bool `json:"pass"`
}

type DiffRequest struct {
	BaseSchema      string `json:"base_schema"`
	HeadSchema      string `json:"head_schema"`
	Config          string `json:"config"`
	SchemaType      string `json:"schema_type"`
	AvroRegistryURL string `json:"avro_registry_url,omitempty"`
	ProviderOrg     string `json:"provider_org,omitempty"`
	ProviderRepo    string `json:"provider_repo,omitempty"`
}

func applyConfig(rep *report.DiffReport, configPath, org, repo string) *report.DiffReport {
	if configPath == "" {
		return rep
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil || cfg == nil {
		return rep
	}

	return diff.ApplyConfigAndTraffic(rep, cfg, org, repo)
}

func runOpenAPIDiff(basePath, headPath, configPath, org, repo string) (*report.DiffReport, error) {
	var rules []config.CustomRule
	if configPath != "" {
		cfg, err := config.LoadConfig(configPath)
		if err == nil && cfg != nil {
			rules = cfg.CustomRules
		}
	}
	rep, err := diff.CompareOpenAPI(basePath, headPath, true, rules)
	if err != nil {
		return nil, err
	}
	rep = applyConfig(rep, configPath, org, repo)
	return rep, nil
}

func runSQLDiff(basePath, headPath, configPath, org, repo string) (*report.DiffReport, error) {
	base, err := sqlpkg.ParseSchema(basePath)
	if err != nil {
		return nil, err
	}
	head, err := sqlpkg.ParseSchema(headPath)
	if err != nil {
		return nil, err
	}
	rep := sqlpkg.DiffSchemas(base, head)
	rep = applyConfig(rep, configPath, org, repo)
	return rep, nil
}

func setupMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/evaluate-gate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal error"}`))
			return
		}
		defer r.Body.Close()

		var req GateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid JSON"}`))
			return
		}

		pass := gates.EvaluateGate(req.Gate, req.BreakingCount, req.WarningCount, req.CrossRepoSafe)

		respBody, err := json.Marshal(GateResponse{Pass: pass})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal error"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
	})

	mux.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal error"}`))
			return
		}
		defer r.Body.Close()

		var req DiffRequest
		if err := json.Unmarshal(body, &req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid JSON"}`))
			return
		}

		if req.BaseSchema == "" || req.HeadSchema == "" || req.SchemaType == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "base_schema, head_schema, and schema_type are required"}`))
			return
		}

		if req.SchemaType != "graphql" && req.SchemaType != "openapi" && req.SchemaType != "sql" && req.SchemaType != "protobuf" && req.SchemaType != "proto" && req.SchemaType != "asyncapi" && req.SchemaType != "avro" && req.SchemaType != "terraform-plan" && req.SchemaType != "ai-model" && req.SchemaType != "aiml" && req.SchemaType != "salesforce-object" && req.SchemaType != "salesforce" && req.SchemaType != "soap-wsdl" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error": "unsupported schema_type: %s"}`, req.SchemaType)))
			return
		}

		ext := ""
		switch req.SchemaType {
		case "graphql":
			ext = ".graphql"
		case "openapi":
			ext = ".yaml"
		case "sql":
			ext = ".sql"
		case "protobuf", "proto":
			ext = ".proto"
		case "avro":
			ext = ".avsc"
		case "aiml", "ai-model":
			ext = ".yaml"
		}

		tempDir, err := os.MkdirTemp("", "substrate-diff-*")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal diff error"}`))
			return
		}
		defer os.RemoveAll(tempDir)

		var baseTarget, headTarget string

		if req.SchemaType == "protobuf" || req.SchemaType == "proto" {
			baseDir := filepath.Join(tempDir, "base")
			headDir := filepath.Join(tempDir, "head")
			os.Mkdir(baseDir, 0755)
			os.Mkdir(headDir, 0755)

			baseFilePath := filepath.Join(baseDir, "schema.proto")
			os.WriteFile(baseFilePath, []byte(req.BaseSchema), 0644)

			headFilePath := filepath.Join(headDir, "schema.proto")
			os.WriteFile(headFilePath, []byte(req.HeadSchema), 0644)

			baseTarget = baseDir
			headTarget = headDir
		} else {
			baseFile, err := os.CreateTemp(tempDir, "base-*"+ext)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal diff error"}`))
				return
			}
			baseFile.WriteString(req.BaseSchema)
			baseFile.Close()
			baseTarget = baseFile.Name()

			headFile, err := os.CreateTemp(tempDir, "head-*"+ext)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal diff error"}`))
				return
			}
			headFile.WriteString(req.HeadSchema)
			headFile.Close()
			headTarget = headFile.Name()
		}

		var configPath string
		if req.Config != "" {
			configFile, err := os.CreateTemp(tempDir, "config-*")
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal diff error"}`))
				return
			}
			configFile.WriteString(req.Config)
			configFile.Close()
			configPath = configFile.Name()
		}

		var rep *report.DiffReport
		switch req.SchemaType {
		case "graphql":
			rep, err = graphql.CompareGraphQL(baseTarget, headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath, req.ProviderOrg, req.ProviderRepo)
			}
		case "sql":
			rep, err = runSQLDiff(baseTarget, headTarget, configPath, req.ProviderOrg, req.ProviderRepo)
		case "terraform-plan":
			rep, err = diff.CompareTerraformPlan(headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath, req.ProviderOrg, req.ProviderRepo)
			}
		case "ai-model", "aiml":
			rep, err = diff.CompareAIML(baseTarget, headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath, req.ProviderOrg, req.ProviderRepo)
			}
		case "asyncapi":
			rep, err = diff.CompareAsyncAPI(baseTarget, headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath, req.ProviderOrg, req.ProviderRepo)
			}
		case "protobuf", "proto":
			rep, err = diff.CompareProto(baseTarget, headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath, req.ProviderOrg, req.ProviderRepo)
			}
		case "avro":
			var cfg *config.SubstrateConfig
			if configPath != "" {
				cfg, _ = config.LoadConfig(configPath)
			}
			rep, err = diff.CompareAvro(baseTarget, headTarget, cfg)
			if err == nil {
				rep = applyConfig(rep, configPath, req.ProviderOrg, req.ProviderRepo)
			}
		case "salesforce-object", "salesforce", "soap-wsdl":
			adapter := &diff.EnterpriseAdapter{}

			baseBytes, readErr := os.ReadFile(baseTarget)
			if readErr != nil {
				err = readErr
				break
			}
			headBytes, readErr := os.ReadFile(headTarget)
			if readErr != nil {
				err = readErr
				break
			}

			rep, err = adapter.Diff(baseBytes, headBytes, nil)
			if err == nil {
				rep = applyConfig(rep, configPath, req.ProviderOrg, req.ProviderRepo)
			}
		default:
			rep, err = runOpenAPIDiff(baseTarget, headTarget, configPath, req.ProviderOrg, req.ProviderRepo)
		}

		if err != nil {
			log.Printf("internal diff error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "internal diff error: %v"}`, err)))
			return
		}

		respBody, err := json.Marshal(rep)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal error"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
	})

	return mux
}

func runServe(port string) error {
	mux := setupMux()
	log.Printf("substrate serve: listening on :%s", port)
	return http.ListenAndServe("0.0.0.0:"+port, mux)
}
