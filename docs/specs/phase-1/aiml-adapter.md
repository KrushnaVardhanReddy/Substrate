# Substrate — Phase 1f: AI/ML Model Contract Diff Adapter

## Status: ⏳ READY

> **Library decision:** No new dependencies needed.
> - `gopkg.in/yaml.v3` — already in `engine/go.mod` — used to parse model contract YAML files.
> - `github.com/santhosh-tekuri/jsonschema/v6` — already in `engine/go.mod` — used to validate dataset schema contracts.
>
> Substrate **owns** the `substrate.yaml` ML model contract format — we define the spec, the diff is pure field-by-field rule comparison. No subprocess, no external service. Single-binary promise maintained.

---

## 1. Architectural Strategy

- **Adapter Location:** `engine/internal/diff/aiml.go`
- **Input Format:** Two `substrate.yaml` files (base vs head) that contain an `ml_model` block.
- **Output:** Standard `report.DiffReport` with `SchemaType` = `"ai-model"`.
- **Schema Type CLI flag:** `--schema-type ai-model`

---

## 2. The `substrate.yaml` ML Model Contract Format

This is the canonical format Substrate defines. Users add this block to their `substrate.yaml` file to declare and protect an ML model's interface contract.

```yaml
# substrate.yaml
service: churn-predictor

ml_model:
  name: churn-predictor
  version: 2.1.0
  framework: pytorch   # pytorch | tensorflow | sklearn | xgboost | onnx | custom
  task: classification # classification | regression | ranking | embedding | generation

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
    input_format: json   # json | csv | parquet
    output_format: json
```

### Go Struct Definitions

The following structs must be defined in `engine/internal/diff/aiml.go`:

```go
type AIMLContract struct {
    Service string     `yaml:"service"`
    MLModel *MLModel   `yaml:"ml_model"`
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
    Name     string      `yaml:"name"`
    Type     string      `yaml:"type"`
    Required bool        `yaml:"required"`
    Range    []float64   `yaml:"range,omitempty"`
    Enum     []string    `yaml:"enum,omitempty"`
}

type Serving struct {
    Endpoint     string `yaml:"endpoint"`
    Method       string `yaml:"method"`
    InputFormat  string `yaml:"input_format"`
    OutputFormat string `yaml:"output_format"`
}
```

---

## 3. Breaking Change Rules

### 3.1 Input Field Rules (Consumer-Side — CRITICAL)

| Rule ID | Severity | Trigger Condition |
|---------|----------|-------------------|
| `AIML_INPUT_REMOVED` | **BREAKING** | A required or optional input field was removed. |
| `AIML_INPUT_TYPE_CHANGED` | **BREAKING** | The `type` of an input field changed (e.g., `float` → `string`). |
| `AIML_INPUT_REQUIRED_ADDED` | **BREAKING** | A new input field was added with `required: true`. Existing consumers cannot provide it. |
| `AIML_INPUT_ENUM_VALUE_REMOVED` | **BREAKING** | A valid enum value was removed from an input field's `enum` list. Existing callers may send that value. |
| `AIML_INPUT_OPTIONAL_ADDED` | **SAFE** | A new optional (`required: false`) input field was added. Backward compatible. |
| `AIML_INPUT_ENUM_VALUE_ADDED` | **SAFE** | A new enum value was added to an input field. |

### 3.2 Output Field Rules (Downstream Pipeline — CRITICAL)

| Rule ID | Severity | Trigger Condition |
|---------|----------|-------------------|
| `AIML_OUTPUT_REMOVED` | **BREAKING** | An output field was removed. Downstream pipelines that parse this field will silently fail. |
| `AIML_OUTPUT_TYPE_CHANGED` | **BREAKING** | The `type` of an output field changed. |
| `AIML_OUTPUT_RANGE_NARROWED` | **WARNING** | The `range` of an output field was narrowed (e.g., `[0.0, 1.0]` → `[0.1, 0.9]`). |
| `AIML_OUTPUT_ENUM_VALUE_REMOVED` | **BREAKING** | A valid enum value was removed from an output field. Downstream parsers may not handle the missing label. |
| `AIML_OUTPUT_ADDED` | **SAFE** | A new output field was added. Backward compatible. |

### 3.3 Model Metadata Rules

| Rule ID | Severity | Trigger Condition |
|---------|----------|-------------------|
| `AIML_TASK_CHANGED` | **BREAKING** | The `task` changed (e.g., `classification` → `regression`). Consumers expecting a probability vector now get a scalar. |
| `AIML_FRAMEWORK_CHANGED` | **WARNING** | The `framework` changed (e.g., `pytorch` → `tensorflow`). May affect serving infrastructure compatibility. |
| `AIML_VERSION_MAJOR_BUMP` | **WARNING** | The major version number increased (e.g., `1.x` → `2.x`). Signals a potentially breaking model update. |

### 3.4 Serving Contract Rules

| Rule ID | Severity | Trigger Condition |
|---------|----------|-------------------|
| `AIML_SERVING_ENDPOINT_CHANGED` | **BREAKING** | The serving `endpoint` path changed. |
| `AIML_SERVING_METHOD_CHANGED` | **BREAKING** | The HTTP `method` changed (e.g., `POST` → `GET`). |
| `AIML_SERVING_FORMAT_CHANGED` | **BREAKING** | The `input_format` or `output_format` changed (e.g., `json` → `csv`). |

---

## 4. Input Validation

- If the YAML file is missing the `ml_model` block entirely, return an error (exit code `3`).
- Parse using `gopkg.in/yaml.v3` into the `AIMLContract` struct.
- If the model `name` field is empty in either base or head, return a parse error.

---

## 5. Diff Function Signature

```go
// CompareAIML reads two substrate.yaml files containing ml_model blocks,
// diffs their contracts, and returns a DiffReport.
func CompareAIML(basePath, headPath string) (*report.DiffReport, error)
```

Sets `SchemaType: "ai-model"` on the returned DiffReport.
