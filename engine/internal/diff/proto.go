package diff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/compliance"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

type bufViolation struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column"`
	Type        string `json:"type"`
	Message     string `json:"message"`
}

func writeBufYAMLIfMissing(dir string) error {
	info, err := os.Stat(dir)
	if err == nil && !info.IsDir() {
		return nil // It's a single file, no buf.yaml needed
	}
	bufYAMLPath := filepath.Join(dir, "buf.yaml")
	if _, err := os.Stat(bufYAMLPath); os.IsNotExist(err) {
		return os.WriteFile(bufYAMLPath, []byte("version: v2\n"), 0644)
	}
	return nil
}

func slugify(s string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(s, "_"))
}

func mapBufType(bufType string) string {
	switch bufType {
	case "FIELD_SAME_TYPE":
		return "PROTO_FIELD_TYPE_CHANGED"
	case "FIELD_SAME_NUMBER":
		return "PROTO_FIELD_NUMBER_CHANGED"
	case "FIELD_SAME_NAME":
		return "PROTO_FIELD_RENAMED"
	case "FIELD_NO_DELETE":
		return "PROTO_FIELD_REMOVED"
	case "FIELD_SAME_LABEL":
		return "PROTO_FIELD_LABEL_CHANGED"
	case "FIELD_SAME_ONEOF":
		return "PROTO_FIELD_ONEOF_CHANGED"
	case "ENUM_NO_DELETE":
		return "PROTO_ENUM_REMOVED"
	case "ENUM_VALUE_NO_DELETE":
		return "PROTO_ENUM_VALUE_REMOVED"
	case "ENUM_VALUE_SAME_NUMBER":
		return "PROTO_ENUM_VALUE_NUMBER_CHANGED"
	case "ENUM_VALUE_SAME_NAME":
		return "PROTO_ENUM_VALUE_RENAMED"
	case "MESSAGE_NO_DELETE":
		return "PROTO_MESSAGE_REMOVED"
	case "RPC_NO_DELETE":
		return "PROTO_RPC_REMOVED"
	case "RPC_SAME_REQUEST_TYPE":
		return "PROTO_RPC_REQUEST_TYPE_CHANGED"
	case "RPC_SAME_RESPONSE_TYPE":
		return "PROTO_RPC_RESPONSE_TYPE_CHANGED"
	case "RPC_SAME_CLIENT_STREAMING", "RPC_SAME_SERVER_STREAMING":
		return "PROTO_RPC_STREAMING_CHANGED"
	case "SERVICE_NO_DELETE":
		return "PROTO_SERVICE_REMOVED"
	case "FILE_SAME_PACKAGE":
		return "PROTO_PACKAGE_CHANGED"
	case "FILE_NO_DELETE":
		return "PROTO_FILE_REMOVED"
	case "PACKAGE_ENUM_NO_DELETE":
		return "PROTO_ENUM_REMOVED"
	case "PACKAGE_MESSAGE_NO_DELETE":
		return "PROTO_MESSAGE_REMOVED"
	case "PACKAGE_SERVICE_NO_DELETE":
		return "PROTO_SERVICE_REMOVED"
	default:
		return "PROTO_" + bufType
	}
}

func mapBufSeverity(bufType string) report.ChangeSeverity {
	switch bufType {
	case "FILE_SAME_PACKAGE":
		return report.ChangeSeverityWarning
	default:
		return report.ChangeSeverityBreaking
	}
}

func recommendationFor(bufType string) string {
	switch mapBufType(bufType) {
	case "PROTO_FIELD_REMOVED":
		return "Do not remove fields. Mark as deprecated instead using `[deprecated=true]`."
	case "PROTO_FIELD_TYPE_CHANGED":
		return "Field type changes break wire compatibility. Add a new field instead."
	case "PROTO_FIELD_NUMBER_CHANGED":
		return "Field numbers are wire-protocol identifiers and must never change."
	case "PROTO_RPC_REMOVED":
		return "Removing an RPC method breaks all clients. Deprecate it first."
	case "PROTO_ENUM_VALUE_REMOVED":
		return "Removing enum values breaks clients that send or receive that value."
	case "PROTO_SERVICE_REMOVED":
		return "Removing a service breaks all clients. Use a deprecation notice first."
	default:
		return "This change breaks wire or source compatibility. Review buf documentation for migration guidance."
	}
}

// CompareProto compares two directories of .proto files and returns a DiffReport.
func CompareProto(baseDir, headDir string) (*report.DiffReport, error) {
	bufPath, err := exec.LookPath("buf")
	if err != nil {
		// Fallback to ~/go/bin/buf in case it's not in PATH
		homeDir, _ := os.UserHomeDir()
		fallbackPath := filepath.Join(homeDir, "go", "bin", "buf")
		if _, statErr := os.Stat(fallbackPath); statErr == nil {
			bufPath = fallbackPath
		} else {
			return nil, fmt.Errorf("buf is not installed or not on PATH — install with: go install github.com/bufbuild/buf/cmd/buf@latest")
		}
	}

	if err := writeBufYAMLIfMissing(baseDir); err != nil {
		return nil, fmt.Errorf("failed to write buf.yaml to baseDir: %w", err)
	}
	if err := writeBufYAMLIfMissing(headDir); err != nil {
		return nil, fmt.Errorf("failed to write buf.yaml to headDir: %w", err)
	}

	cmd := exec.Command(bufPath, "breaking", headDir, "--against", baseDir, "--error-format", "json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run() // exit code 1 can mean either buf error or breaking changes

	// If stderr has text but stdout is empty, it's a buf error
	if stderr.Len() > 0 && stdout.Len() == 0 {
		return nil, fmt.Errorf("buf error: %s", stderr.String())
	}

	rep := &report.DiffReport{
		SubstrateVersion: "0.1.0",
		SchemaType:       report.SchemaTypeProtobuf,
		ComparedAt:       time.Now().UTC().Format(time.RFC3339),
		BreakingChanges:  []report.Change{},
		Warnings:         []report.Change{},
		SafeChanges:      []report.Change{},
	}

	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var violation bufViolation
		if err := json.Unmarshal([]byte(line), &violation); err != nil {
			return nil, fmt.Errorf("failed to parse buf output: %w", err)
		}

		rec := recommendationFor(violation.Type)

		relPath := violation.Path
		if r, err := filepath.Rel(headDir, violation.Path); err == nil {
			relPath = r
		}

		change := report.Change{
			ID:             fmt.Sprintf("proto_%s_%s_%d", strings.ToLower(violation.Type), slugify(relPath), violation.StartLine),
			RuleID:         mapBufType(violation.Type),
			Severity:       mapBufSeverity(violation.Type),
			Path:           fmt.Sprintf("%s:%d", relPath, violation.StartLine),
			Description:    violation.Message,
			Recommendation: &rec,
		}

		switch change.Severity {
		case report.ChangeSeverityBreaking:
			rep.BreakingChanges = append(rep.BreakingChanges, change)
		case report.ChangeSeverityWarning:
			rep.Warnings = append(rep.Warnings, change)
		default:
			rep.SafeChanges = append(rep.SafeChanges, change)
		}
	}

	rep.Summary.BreakingCount = len(rep.BreakingChanges)
	rep.Summary.WarningCount = len(rep.Warnings)
	rep.Summary.SafeCount = len(rep.SafeChanges)
	rep.Summary.TotalChanges = rep.Summary.BreakingCount + rep.Summary.WarningCount + rep.Summary.SafeCount

	if rep.Summary.BreakingCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityBreaking
	} else if rep.Summary.WarningCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityWarning
	} else if rep.Summary.SafeCount > 0 {
		rep.Summary.OverallSeverity = report.SeveritySafe
	} else {
		rep.Summary.OverallSeverity = report.SeverityNoChanges
	}

	compliance.Audit(rep)

	return rep, nil
}
