package diff

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/compliance"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"gopkg.in/yaml.v3"
)

type AIMLContract struct {
	Service string   `yaml:"service"`
	MLModel *MLModel `yaml:"ml_model"`
}

type MLModel struct {
	Name      string       `yaml:"name"`
	Version   string       `yaml:"version"`
	Framework string       `yaml:"framework"`
	Task      string       `yaml:"task"`
	Inputs    []ModelField `yaml:"inputs"`
	Outputs   []ModelField `yaml:"outputs"`
	Serving   *Serving     `yaml:"serving,omitempty"`
}

type ModelField struct {
	Name     string    `yaml:"name"`
	Type     string    `yaml:"type"`
	Required bool      `yaml:"required"`
	Range    []float64 `yaml:"range,omitempty"`
	Enum     []string  `yaml:"enum,omitempty"`
}

type Serving struct {
	Endpoint     string `yaml:"endpoint"`
	Method       string `yaml:"method"`
	InputFormat  string `yaml:"input_format"`
	OutputFormat string `yaml:"output_format"`
}

func parseAIMLContract(path string) (*AIMLContract, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var contract AIMLContract
	if err := yaml.Unmarshal(bytes, &contract); err != nil {
		return nil, err
	}

	return &contract, nil
}

func parseMajorVersion(version string) int {
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return 0
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	return major
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func CompareAIML(basePath, headPath string) (*report.DiffReport, error) {
	baseContract, err := parseAIMLContract(basePath)
	if err != nil {
		return nil, err
	}
	headContract, err := parseAIMLContract(headPath)
	if err != nil {
		return nil, err
	}

	if baseContract.MLModel == nil {
		return nil, fmt.Errorf("no ml_model block found in %s", basePath)
	}
	if headContract.MLModel == nil {
		return nil, fmt.Errorf("no ml_model block found in %s", headPath)
	}

	base := baseContract.MLModel
	head := headContract.MLModel

	if base.Name == "" {
		return nil, fmt.Errorf("empty ml_model name in %s", basePath)
	}
	if head.Name == "" {
		return nil, fmt.Errorf("empty ml_model name in %s", headPath)
	}

	rep := &report.DiffReport{
		SchemaType:      "ai-model",
		ComparedAt:      time.Now().UTC().Format(time.RFC3339),
		BreakingChanges: make([]report.Change, 0),
		Warnings:        make([]report.Change, 0),
		SafeChanges:     make([]report.Change, 0),
	}

	// Diff Metadata
	if base.Task != head.Task {
		rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
			ID:          "aiml-task-changed",
			RuleID:      "AIML_TASK_CHANGED",
			Severity:    report.ChangeSeverityBreaking,
			Path:        "ml_model.task",
			Description: fmt.Sprintf("Task changed from '%s' to '%s'", base.Task, head.Task),
			Before:      base.Task,
			After:       head.Task,
		})
	}

	if base.Framework != head.Framework {
		rep.Warnings = append(rep.Warnings, report.Change{
			ID:          "aiml-framework-changed",
			RuleID:      "AIML_FRAMEWORK_CHANGED",
			Severity:    report.ChangeSeverityWarning,
			Path:        "ml_model.framework",
			Description: fmt.Sprintf("Framework changed from '%s' to '%s'", base.Framework, head.Framework),
			Before:      base.Framework,
			After:       head.Framework,
		})
	}

	baseMajor := parseMajorVersion(base.Version)
	headMajor := parseMajorVersion(head.Version)
	if headMajor > baseMajor {
		rep.Warnings = append(rep.Warnings, report.Change{
			ID:          "aiml-version-major-bump",
			RuleID:      "AIML_VERSION_MAJOR_BUMP",
			Severity:    report.ChangeSeverityWarning,
			Path:        "ml_model.version",
			Description: fmt.Sprintf("Major version bumped from '%s' to '%s'", base.Version, head.Version),
			Before:      base.Version,
			After:       head.Version,
		})
	}

	// Diff Inputs
	baseInputs := make(map[string]ModelField)
	for _, in := range base.Inputs {
		baseInputs[in.Name] = in
	}

	headInputs := make(map[string]ModelField)
	for _, in := range head.Inputs {
		headInputs[in.Name] = in
	}

	for name, baseIn := range baseInputs {
		headIn, ok := headInputs[name]
		if !ok {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("aiml-input-removed-%s", name),
				RuleID:      "AIML_INPUT_REMOVED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("ml_model.inputs.%s", name),
				Description: fmt.Sprintf("Input field '%s' was removed", name),
				Before:      name,
				After:       nil,
			})
			continue
		}

		if baseIn.Type != headIn.Type {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("aiml-input-type-changed-%s", name),
				RuleID:      "AIML_INPUT_TYPE_CHANGED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("ml_model.inputs.%s.type", name),
				Description: fmt.Sprintf("Type of input '%s' changed from '%s' to '%s'", name, baseIn.Type, headIn.Type),
				Before:      baseIn.Type,
				After:       headIn.Type,
			})
		}

		for _, baseEnumVal := range baseIn.Enum {
			if !containsString(headIn.Enum, baseEnumVal) {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:          fmt.Sprintf("aiml-input-enum-removed-%s-%s", name, baseEnumVal),
					RuleID:      "AIML_INPUT_ENUM_VALUE_REMOVED",
					Severity:    report.ChangeSeverityBreaking,
					Path:        fmt.Sprintf("ml_model.inputs.%s.enum", name),
					Description: fmt.Sprintf("Enum value '%s' removed from input '%s'", baseEnumVal, name),
					Before:      baseEnumVal,
					After:       nil,
				})
			}
		}

		for _, headEnumVal := range headIn.Enum {
			if !containsString(baseIn.Enum, headEnumVal) {
				rep.SafeChanges = append(rep.SafeChanges, report.Change{
					ID:          fmt.Sprintf("aiml-input-enum-added-%s-%s", name, headEnumVal),
					RuleID:      "AIML_INPUT_ENUM_VALUE_ADDED",
					Severity:    report.ChangeSeveritySafe,
					Path:        fmt.Sprintf("ml_model.inputs.%s.enum", name),
					Description: fmt.Sprintf("Enum value '%s' added to input '%s'", headEnumVal, name),
					Before:      nil,
					After:       headEnumVal,
				})
			}
		}
	}

	for name, headIn := range headInputs {
		if _, ok := baseInputs[name]; !ok {
			if headIn.Required {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:          fmt.Sprintf("aiml-input-req-added-%s", name),
					RuleID:      "AIML_INPUT_REQUIRED_ADDED",
					Severity:    report.ChangeSeverityBreaking,
					Path:        fmt.Sprintf("ml_model.inputs.%s", name),
					Description: fmt.Sprintf("Required input field '%s' was added", name),
					Before:      nil,
					After:       name,
				})
			} else {
				rep.SafeChanges = append(rep.SafeChanges, report.Change{
					ID:          fmt.Sprintf("aiml-input-opt-added-%s", name),
					RuleID:      "AIML_INPUT_OPTIONAL_ADDED",
					Severity:    report.ChangeSeveritySafe,
					Path:        fmt.Sprintf("ml_model.inputs.%s", name),
					Description: fmt.Sprintf("Optional input field '%s' was added", name),
					Before:      nil,
					After:       name,
				})
			}
		}
	}

	// Diff Outputs
	baseOutputs := make(map[string]ModelField)
	for _, out := range base.Outputs {
		baseOutputs[out.Name] = out
	}

	headOutputs := make(map[string]ModelField)
	for _, out := range head.Outputs {
		headOutputs[out.Name] = out
	}

	for name, baseOut := range baseOutputs {
		headOut, ok := headOutputs[name]
		if !ok {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("aiml-output-removed-%s", name),
				RuleID:      "AIML_OUTPUT_REMOVED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("ml_model.outputs.%s", name),
				Description: fmt.Sprintf("Output field '%s' was removed", name),
				Before:      name,
				After:       nil,
			})
			continue
		}

		if baseOut.Type != headOut.Type {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("aiml-output-type-changed-%s", name),
				RuleID:      "AIML_OUTPUT_TYPE_CHANGED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("ml_model.outputs.%s.type", name),
				Description: fmt.Sprintf("Type of output '%s' changed from '%s' to '%s'", name, baseOut.Type, headOut.Type),
				Before:      baseOut.Type,
				After:       headOut.Type,
			})
		}

		for _, baseEnumVal := range baseOut.Enum {
			if !containsString(headOut.Enum, baseEnumVal) {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:          fmt.Sprintf("aiml-output-enum-removed-%s-%s", name, baseEnumVal),
					RuleID:      "AIML_OUTPUT_ENUM_VALUE_REMOVED",
					Severity:    report.ChangeSeverityBreaking,
					Path:        fmt.Sprintf("ml_model.outputs.%s.enum", name),
					Description: fmt.Sprintf("Enum value '%s' removed from output '%s'", baseEnumVal, name),
					Before:      baseEnumVal,
					After:       nil,
				})
			}
		}

		if len(baseOut.Range) == 2 && len(headOut.Range) == 2 {
			if headOut.Range[0] > baseOut.Range[0] || headOut.Range[1] < baseOut.Range[1] {
				rep.Warnings = append(rep.Warnings, report.Change{
					ID:          fmt.Sprintf("aiml-output-range-narrowed-%s", name),
					RuleID:      "AIML_OUTPUT_RANGE_NARROWED",
					Severity:    report.ChangeSeverityWarning,
					Path:        fmt.Sprintf("ml_model.outputs.%s.range", name),
					Description: fmt.Sprintf("Range of output '%s' narrowed from [%v, %v] to [%v, %v]", name, baseOut.Range[0], baseOut.Range[1], headOut.Range[0], headOut.Range[1]),
					Before:      baseOut.Range,
					After:       headOut.Range,
				})
			}
		}
	}

	for name := range headOutputs {
		if _, ok := baseOutputs[name]; !ok {
			rep.SafeChanges = append(rep.SafeChanges, report.Change{
				ID:          fmt.Sprintf("aiml-output-added-%s", name),
				RuleID:      "AIML_OUTPUT_ADDED",
				Severity:    report.ChangeSeveritySafe,
				Path:        fmt.Sprintf("ml_model.outputs.%s", name),
				Description: fmt.Sprintf("Output field '%s' was added", name),
				Before:      nil,
				After:       name,
			})
		}
	}

	// Diff Serving
	if base.Serving != nil && head.Serving != nil {
		if base.Serving.Endpoint != head.Serving.Endpoint {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          "aiml-serving-endpoint-changed",
				RuleID:      "AIML_SERVING_ENDPOINT_CHANGED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        "ml_model.serving.endpoint",
				Description: fmt.Sprintf("Serving endpoint changed from '%s' to '%s'", base.Serving.Endpoint, head.Serving.Endpoint),
				Before:      base.Serving.Endpoint,
				After:       head.Serving.Endpoint,
			})
		}
		if base.Serving.Method != head.Serving.Method {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          "aiml-serving-method-changed",
				RuleID:      "AIML_SERVING_METHOD_CHANGED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        "ml_model.serving.method",
				Description: fmt.Sprintf("Serving method changed from '%s' to '%s'", base.Serving.Method, head.Serving.Method),
				Before:      base.Serving.Method,
				After:       head.Serving.Method,
			})
		}
		if base.Serving.InputFormat != head.Serving.InputFormat {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          "aiml-serving-input-format-changed",
				RuleID:      "AIML_SERVING_FORMAT_CHANGED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        "ml_model.serving.input_format",
				Description: fmt.Sprintf("Input format changed from '%s' to '%s'", base.Serving.InputFormat, head.Serving.InputFormat),
				Before:      base.Serving.InputFormat,
				After:       head.Serving.InputFormat,
			})
		}
		if base.Serving.OutputFormat != head.Serving.OutputFormat {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          "aiml-serving-output-format-changed",
				RuleID:      "AIML_SERVING_FORMAT_CHANGED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        "ml_model.serving.output_format",
				Description: fmt.Sprintf("Output format changed from '%s' to '%s'", base.Serving.OutputFormat, head.Serving.OutputFormat),
				Before:      base.Serving.OutputFormat,
				After:       head.Serving.OutputFormat,
			})
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
