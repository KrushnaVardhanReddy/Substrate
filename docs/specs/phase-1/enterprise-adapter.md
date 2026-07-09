# Phase 1g: Enterprise Metadata (Salesforce & SOAP)

## 🎯 Overview
This adapter completes the Substrate engine by bringing legacy enterprise systems (Salesforce metadata, SOAP WSDLs) into the modern CI/CD orchestration layer. This allows a company to automatically block a PR if changing a Salesforce Custom Object breaks a downstream GraphQL API, or if altering a WSDL breaks a legacy payment integration.

**Owner:** Jules  
**Status:** ⏳ READY  

---

## 1. Adapter Architecture

The Enterprise adapter (`engine/internal/diff/enterprise.go`) will parse two distinct formats via XML:
1. **Salesforce Metadata API XML:** Primarily `.object` (Custom Objects) and `.profile` (Field Level Security).
2. **SOAP WSDL:** WSDL 1.1 definitions.

Since both are heavily XML-based, we will rely on Go's `encoding/xml`.

---

## 2. Salesforce Custom Object Rules (`.object`)

A Salesforce object XML typically looks like:
```xml
<CustomObject xmlns="http://soap.sforce.com/2006/04/metadata">
    <fields>
        <fullName>Customer_ID__c</fullName>
        <type>Text</type>
        <required>true</required>
    </fields>
</CustomObject>
```

| ID | Description | Severity | Why it matters |
|---|---|---|---|
| `SFDC_FIELD_REMOVED` | A custom field (`<fields>`) is deleted. | **BREAKING** | Integrations syncing to this field will fail (e.g. MuleSoft/Boomi integrations). |
| `SFDC_FIELD_REQUIRED` | An existing field changes `<required>false</required>` to `true`. | **BREAKING** | Upstream systems POSTing records without this field will now get `REQUIRED_FIELD_MISSING` errors. |
| `SFDC_FIELD_TYPE_CHANGED` | `<type>` changes (e.g., `Text` to `Number`). | **BREAKING** | Causes SOQL/Apex cast exceptions in consumers. |
| `SFDC_FIELD_LENGTH_REDUCED` | `<length>` of a text field is reduced (e.g., 255 to 100). | **BREAKING** | Data truncation errors during ingestion. |
| `SFDC_FIELD_ADDED` | A new field is added. | Safe | Upstream systems can safely ignore it. |

---

## 3. SOAP WSDL Rules (`.wsdl`)

WSDLs are the lifeblood of enterprise integrations (banks, ERPs).

| ID | Description | Severity | Why it matters |
|---|---|---|---|
| `WSDL_OPERATION_REMOVED` | An `<operation>` inside a `<portType>` is removed. | **BREAKING** | Consumer systems can no longer call this RPC method. |
| `WSDL_INPUT_MESSAGE_CHANGED` | The `message` attribute of an `<input>` is altered. | **BREAKING** | The required request payload schema has changed. |
| `WSDL_FAULT_ADDED` | A new `<fault>` is added to an operation. | **WARNING** | Consumers might not have exception handling for the new fault type. |
| `WSDL_BINDING_REMOVED` | A `<binding>` is deleted. | **BREAKING** | The protocol (e.g., SOAP 1.2 vs 1.1 over HTTP) is no longer supported. |

---

## 4. `substrate.yaml` Configuration

```yaml
version: 1.0.0
contracts:
  - type: salesforce-object
    path: src/objects/Account.object
  - type: soap-wsdl
    path: src/wsdl/PaymentService.wsdl
```

---

## 5. Go Interface Implementation Requirements

The adapter must implement `engine/internal/diff/DiffEngine`.

**Struct Definition:**
```go
package diff

import (
	"encoding/xml"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

type EnterpriseAdapter struct{}

func (a *EnterpriseAdapter) Diff(base, head []byte, config map[string]interface{}) (*report.DiffReport, error) {
    // Detect if XML is WSDL (contains <definitions>) or SFDC (<CustomObject>)
    // and route to respective diffing logic.
}
```

*Note: Since XML parsing in Go requires strictly typed structs, the adapter must define internal `structs` with `xml:"..."` tags for `CustomObject`, `CustomField`, `Definitions`, `PortType`, and `Operation`.*
