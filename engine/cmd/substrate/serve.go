package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	sqlpkg "github.com/KrushnaVardhanReddy/substrate/engine/internal/sql"
)

type DiffRequest struct {
	BaseSchema      string `json:"base_schema"`
	HeadSchema      string `json:"head_schema"`
	Config          string `json:"config"`
	SchemaType      string `json:"schema_type"`
	AvroRegistryURL string `json:"avro_registry_url,omitempty"`
}

func applyConfig(rep *report.DiffReport, configPath string) *report.DiffReport {
	if configPath == "" {
		return rep
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil || cfg == nil {
		return rep
	}

	var activeBreaking []report.Change
	for _, bc := range rep.BreakingChanges {
		if cfg.IsOverrideActive(bc.RuleID, bc.Path) {
			continue
		}
		activeBreaking = append(activeBreaking, bc)
	}

	rep.BreakingChanges = activeBreaking
	rep.Summary.BreakingCount = len(rep.BreakingChanges)

	if rep.Summary.BreakingCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityBreaking
	} else if rep.Summary.WarningCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityWarning
	} else if rep.Summary.TotalChanges > 0 {
		rep.Summary.OverallSeverity = report.SeveritySafe
	} else {
		rep.Summary.OverallSeverity = report.SeverityNoChanges
	}

	return rep
}

func runOpenAPIDiff(basePath, headPath, configPath string) (*report.DiffReport, error) {
	rep, err := diff.CompareOpenAPI(basePath, headPath, true)
	if err != nil {
		return nil, err
	}
	rep = applyConfig(rep, configPath)
	return rep, nil
}

func runSQLDiff(basePath, headPath, configPath string) (*report.DiffReport, error) {
	base, err := sqlpkg.ParseSchema(basePath)
	if err != nil {
		return nil, err
	}
	head, err := sqlpkg.ParseSchema(headPath)
	if err != nil {
		return nil, err
	}
	rep := sqlpkg.DiffSchemas(base, head)
	rep = applyConfig(rep, configPath)
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

		if req.SchemaType != "openapi" && req.SchemaType != "sql" && req.SchemaType != "protobuf" && req.SchemaType != "proto" && req.SchemaType != "asyncapi" && req.SchemaType != "avro" && req.SchemaType != "terraform-plan" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error": "unsupported schema_type: %s"}`, req.SchemaType)))
			return
		}

		baseFile, err := os.CreateTemp("", "base-*")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal diff error"}`))
			return
		}
		defer os.Remove(baseFile.Name())
		baseFile.WriteString(req.BaseSchema)
		baseFile.Close()

		headFile, err := os.CreateTemp("", "head-*")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal diff error"}`))
			return
		}
		defer os.Remove(headFile.Name())
		headFile.WriteString(req.HeadSchema)
		headFile.Close()

		var configPath string
		if req.Config != "" {
			configFile, err := os.CreateTemp("", "config-*")
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal diff error"}`))
				return
			}
			defer os.Remove(configFile.Name())
			configFile.WriteString(req.Config)
			configFile.Close()
			configPath = configFile.Name()
		}

		var rep *report.DiffReport
		switch req.SchemaType {
		case "sql":
			rep, err = runSQLDiff(baseFile.Name(), headFile.Name(), configPath)
		case "terraform-plan":
			rep, err = diff.CompareTerraformPlan(headFile.Name())
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		case "asyncapi":
			rep, err = diff.CompareAsyncAPI(baseFile.Name(), headFile.Name())
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		case "protobuf", "proto":
			rep, err = diff.CompareProto(baseFile.Name(), headFile.Name())
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		case "avro":
			avroCfg := &config.SubstrateConfig{
				Service: "unknown",
				Avro: &config.AvroConfig{
					SchemaRegistryURL: req.AvroRegistryURL,
				},
			}
			rep, err = diff.CompareAvro(baseFile.Name(), headFile.Name(), avroCfg)
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		default:
			rep, err = runOpenAPIDiff(baseFile.Name(), headFile.Name(), configPath)
		}

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal diff error"}`))
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
