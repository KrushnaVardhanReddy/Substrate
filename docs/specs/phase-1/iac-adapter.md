# Phase 1h: Infrastructure as Code (Terraform) Adapter

## Overview
Substrate diffs API schemas (OpenAPI, SQL, Protobuf, AsyncAPI) to prevent breaking changes. However, API contracts often remain unchanged while the underlying infrastructure (e.g., DynamoDB tables, S3 buckets, SNS topics) is modified or destroyed via Terraform.

The Terraform adapter parses the JSON output of `terraform plan` (`terraform show -json tfplan`) to detect breaking changes in infrastructure state. 

## Engine Integration

### CLI Inputs
Terraform does not use the standard `base_schema` vs `head_schema` pattern, as a `tfplan` natively contains the diff.
When the user specifies `--schema-type terraform-plan`, Substrate will treat the `--head` argument as the `tfplan.json` file and ignore the `--base` argument.

### Parser Dependency
We use Go's standard library `encoding/json` to unmarshal the `terraform plan` JSON output. We do not need the full `hashicorp/terraform-json` struct, just enough to extract `resource_changes`, `output_changes`, and `variables`.

```go
type TFPlan struct {
	ResourceChanges []TFResourceChange `json:"resource_changes"`
	OutputChanges   map[string]TFOutputChange `json:"output_changes"`
}
```

## Breaking Change Rules

| Rule ID | Severity | Description | Trigger Condition |
|---------|----------|-------------|-------------------|
| `TF_RESOURCE_DESTROYED` | BREAKING | A stateful resource is being destroyed | `resource_change.actions` contains `"delete"` without `"create"` (or `"delete"` then `"create"` without state migration). We flag *any* delete action for stateful types (e.g., `aws_dynamodb_table`, `aws_s3_bucket`, `aws_db_instance`). |
| `TF_OUTPUT_REMOVED` | BREAKING | A Terraform output is removed | `output_changes[name].actions` contains `"delete"` |
| `TF_IAM_PERMISSION_REMOVED` | WARNING | An IAM policy allows fewer actions | Emit a warning if `aws_iam_policy` or `aws_iam_role_policy` has an `"update"` action. |

## Usage
Users will integrate this in CI by generating a plan and passing the JSON to Substrate:
```bash
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
substrate diff /dev/null tfplan.json --schema-type terraform-plan
```
