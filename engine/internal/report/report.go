package report

type SchemaType string

const (
	SchemaTypeOpenAPI  SchemaType = "openapi"
	SchemaTypeSQL      SchemaType = "sql"
	SchemaTypeGraphQL  SchemaType = "graphql"
	SchemaTypeProtobuf SchemaType = "protobuf"
	SchemaTypeAvro     SchemaType = "avro"
)

type Severity string

const (
	SeverityBreaking  Severity = "BREAKING"
	SeverityWarning   Severity = "WARNING"
	SeveritySafe      Severity = "SAFE"
	SeverityNoChanges Severity = "NO_CHANGES"
)

type ChangeSeverity string

const (
	ChangeSeverityBreaking ChangeSeverity = "BREAKING"
	ChangeSeverityWarning  ChangeSeverity = "WARNING"
	ChangeSeveritySafe     ChangeSeverity = "SAFE"
)

type Summary struct {
	TotalChanges    int      `json:"total_changes"`
	BreakingCount   int      `json:"breaking_count"`
	WarningCount    int      `json:"warning_count"`
	SafeCount       int      `json:"safe_count"`
	OverallSeverity Severity `json:"overall_severity"`
}

type Change struct {
	ID              string         `json:"id"`
	RuleID          string         `json:"rule_id"`
	Severity        ChangeSeverity `json:"severity"`
	Path            string         `json:"path"`
	Description     string         `json:"description"`
	Before          any            `json:"before"`
	After           any            `json:"after"`
	Recommendation  *string        `json:"recommendation"`
	ConsumerImpacts []string       `json:"consumer_impacts,omitempty"`
}

type ComplianceAlert struct {
	Path           string `json:"path"`
	ComplianceType string `json:"compliance_type"`
	Message        string `json:"message"`
}

type DiffReport struct {
	SubstrateVersion     string                `json:"substrate_version"`
	SchemaType           SchemaType            `json:"schema_type"`
	ComparedAt           string                `json:"compared_at"`
	Mode                 string                `json:"mode"`
	Summary              Summary               `json:"summary"`
	BreakingChanges      []Change              `json:"breaking_changes"`
	Warnings             []Change              `json:"warnings"`
	SafeChanges          []Change              `json:"safe_changes"`
	ComplianceAlerts     []ComplianceAlert     `json:"compliance_alerts,omitempty"`
	Deprecations         []Deprecation         `json:"deprecations,omitempty"`
	GovernanceViolations []GovernanceViolation `json:"governance_violations,omitempty"`
}

type Deprecation struct {
	Endpoint   string `json:"endpoint"`
	SunsetDate string `json:"sunset_date"`
}

type GovernanceViolation struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}
