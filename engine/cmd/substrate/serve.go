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

		if req.SchemaType != "graphql" && req.SchemaType != "openapi" && req.SchemaType != "sql" && req.SchemaType != "protobuf" && req.SchemaType != "proto" && req.SchemaType != "asyncapi" && req.SchemaType != "avro" && req.SchemaType != "terraform-plan" && req.SchemaType != "ai-model" {
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
			rep, err = diff.CompareGraphQL(baseTarget, headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		case "sql":
			rep, err = runSQLDiff(baseTarget, headTarget, configPath)
		case "terraform-plan":
			rep, err = diff.CompareTerraformPlan(headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		case "ai-model":
			rep, err = diff.CompareAIML(baseTarget, headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		case "asyncapi":
			rep, err = diff.CompareAsyncAPI(baseTarget, headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		case "protobuf", "proto":
			rep, err = diff.CompareProto(baseTarget, headTarget)
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		case "avro":
			var cfg *config.SubstrateConfig
			if configPath != "" {
				cfg, _ = config.LoadConfig(configPath)
			}
			rep, err = diff.CompareAvro(baseTarget, headTarget, cfg)
			if err == nil {
				rep = applyConfig(rep, configPath)
			}
		default:
			rep, err = runOpenAPIDiff(baseTarget, headTarget, configPath)
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
