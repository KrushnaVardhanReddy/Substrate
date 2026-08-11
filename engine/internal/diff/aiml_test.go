// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package diff_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
)

func TestCompareAIML(t *testing.T) {
	baseYAML := `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`

	tests := []struct {
		name              string
		headYAML          string
		expectError       bool
		expectedBreaking  int
		expectedWarning   int
		expectedSafe      int
		expectedRuleBreak string
		expectedRuleWarn  string
		expectedRuleSafe  string
	}{
		{
			name:             "no changes",
			headYAML:         baseYAML,
			expectError:      false,
			expectedBreaking: 0,
			expectedWarning:  0,
			expectedSafe:     0,
		},
		{
			name: "input removed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_INPUT_REMOVED",
		},
		{
			name: "input type changed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: string
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_INPUT_TYPE_CHANGED",
		},
		{
			name: "required input added",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
    - name: new_feature
      type: float
      required: true
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_INPUT_REQUIRED_ADDED",
		},
		{
			name: "optional input added",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
    - name: extra_flag
      type: bool
      required: false
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:      false,
			expectedBreaking: 0,
			expectedSafe:     1,
			expectedRuleSafe: "AIML_INPUT_OPTIONAL_ADDED",
		},
		{
			name: "input enum value removed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_INPUT_ENUM_VALUE_REMOVED",
		},
		{
			name: "output removed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_OUTPUT_REMOVED",
		},
		{
			name: "output type changed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: int
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_OUTPUT_TYPE_CHANGED",
		},
		{
			name: "output range narrowed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.1, 0.9]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:      false,
			expectedBreaking: 0,
			expectedWarning:  1,
			expectedRuleWarn: "AIML_OUTPUT_RANGE_NARROWED",
		},
		{
			name: "output enum value removed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_OUTPUT_ENUM_VALUE_REMOVED",
		},
		{
			name: "task changed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: regression
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_TASK_CHANGED",
		},
		{
			name: "framework changed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: tensorflow
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:      false,
			expectedBreaking: 0,
			expectedWarning:  1,
			expectedRuleWarn: "AIML_FRAMEWORK_CHANGED",
		},
		{
			name: "version major bump",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 2.0.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:      false,
			expectedBreaking: 0,
			expectedWarning:  1,
			expectedRuleWarn: "AIML_VERSION_MAJOR_BUMP",
		},
		{
			name: "missing ml_model block",
			headYAML: `
service: churn-predictor
version: 1.5.0
`,
			expectError: true,
		},
		{
			name: "serving endpoint changed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /v2/predict
    method: POST
    input_format: json
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_SERVING_ENDPOINT_CHANGED",
		},
		{
			name: "serving format changed",
			headYAML: `
service: churn-predictor
ml_model:
  name: churn-predictor
  version: 1.5.0
  framework: pytorch
  task: classification
  inputs:
    - name: tenure_months
      type: float
      required: true
    - name: monthly_charges
      type: float
      required: true
    - name: contract_type
      type: string
      required: true
      enum: [Month-to-month, One year, Two year]
  outputs:
    - name: churn_probability
      type: float
      range: [0.0, 1.0]
    - name: churn_label
      type: string
      enum: [Yes, No]
  serving:
    endpoint: /predict
    method: POST
    input_format: csv
    output_format: json
`,
			expectError:       false,
			expectedBreaking:  1,
			expectedRuleBreak: "AIML_SERVING_FORMAT_CHANGED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseFile := filepath.Join(t.TempDir(), "base.yaml")
			headFile := filepath.Join(t.TempDir(), "head.yaml")

			if err := os.WriteFile(baseFile, []byte(baseYAML), 0644); err != nil {
				t.Fatalf("failed to write base file: %v", err)
			}
			if err := os.WriteFile(headFile, []byte(tt.headYAML), 0644); err != nil {
				t.Fatalf("failed to write head file: %v", err)
			}

			rep, err := diff.CompareAIML(baseFile, headFile)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected an error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rep.Summary.BreakingCount != tt.expectedBreaking {
				t.Errorf("expected %d breaking changes, got %d", tt.expectedBreaking, rep.Summary.BreakingCount)
			}
			if rep.Summary.WarningCount != tt.expectedWarning {
				t.Errorf("expected %d warnings, got %d", tt.expectedWarning, rep.Summary.WarningCount)
			}
			if rep.Summary.SafeCount != tt.expectedSafe {
				t.Errorf("expected %d safe changes, got %d", tt.expectedSafe, rep.Summary.SafeCount)
			}

			if tt.expectedRuleBreak != "" {
				found := false
				for _, bc := range rep.BreakingChanges {
					if bc.RuleID == tt.expectedRuleBreak {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected breaking rule %s but not found", tt.expectedRuleBreak)
				}
			}

			if tt.expectedRuleWarn != "" {
				found := false
				for _, w := range rep.Warnings {
					if w.RuleID == tt.expectedRuleWarn {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning rule %s but not found", tt.expectedRuleWarn)
				}
			}

			if tt.expectedRuleSafe != "" {
				found := false
				for _, sc := range rep.SafeChanges {
					if sc.RuleID == tt.expectedRuleSafe {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected safe rule %s but not found", tt.expectedRuleSafe)
				}
			}
		})
	}
}
