// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package diff

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

func (a *EnterpriseAdapter) diffSalesforce(base, head []byte) (*report.DiffReport, error) {
	var baseObj, headObj CustomObject

	if err := xml.Unmarshal(base, &baseObj); err != nil {
		return nil, fmt.Errorf("failed to parse base Salesforce object: %w", err)
	}

	if err := xml.Unmarshal(head, &headObj); err != nil {
		return nil, fmt.Errorf("failed to parse head Salesforce object: %w", err)
	}

	rep := &report.DiffReport{
		SubstrateVersion: "0.1.0",
		SchemaType:       "salesforce",
		ComparedAt:       time.Now().UTC().Format(time.RFC3339),
		Summary: report.Summary{
			OverallSeverity: report.SeverityNoChanges,
		},
		BreakingChanges: []report.Change{},
		Warnings:        []report.Change{},
		SafeChanges:     []report.Change{},
	}

	baseFields := make(map[string]CustomField)
	for _, f := range baseObj.Fields {
		baseFields[f.FullName] = f
	}

	headFields := make(map[string]CustomField)
	for _, f := range headObj.Fields {
		headFields[f.FullName] = f
	}

	for name, baseF := range baseFields {
		headF, ok := headFields[name]
		if !ok {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("sfdc-field-removed-%s", name),
				RuleID:      "SFDC_FIELD_REMOVED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("fields.%s", name),
				Description: fmt.Sprintf("Salesforce field '%s' was removed", name),
				Before:      name,
				After:       nil,
			})
			continue
		}

		if !baseF.Required && headF.Required {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("sfdc-field-required-%s", name),
				RuleID:      "SFDC_FIELD_REQUIRED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("fields.%s.required", name),
				Description: fmt.Sprintf("Field '%s' was made required", name),
				Before:      baseF.Required,
				After:       headF.Required,
			})
		}

		if baseF.Type != headF.Type {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("sfdc-field-type-changed-%s", name),
				RuleID:      "SFDC_FIELD_TYPE_CHANGED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("fields.%s.type", name),
				Description: fmt.Sprintf("Type of field '%s' changed from '%s' to '%s'", name, baseF.Type, headF.Type),
				Before:      baseF.Type,
				After:       headF.Type,
			})
		}

		if baseF.Length > 0 && headF.Length > 0 && headF.Length < baseF.Length {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("sfdc-field-length-reduced-%s", name),
				RuleID:      "SFDC_FIELD_LENGTH_REDUCED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("fields.%s.length", name),
				Description: fmt.Sprintf("Length of field '%s' reduced from %d to %d", name, baseF.Length, headF.Length),
				Before:      baseF.Length,
				After:       headF.Length,
			})
		}
	}

	for name := range headFields {
		if _, ok := baseFields[name]; !ok {
			rep.SafeChanges = append(rep.SafeChanges, report.Change{
				ID:          fmt.Sprintf("sfdc-field-added-%s", name),
				RuleID:      "SFDC_FIELD_ADDED",
				Severity:    report.ChangeSeveritySafe,
				Path:        fmt.Sprintf("fields.%s", name),
				Description: fmt.Sprintf("Salesforce field '%s' was added", name),
				Before:      nil,
				After:       name,
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

	return rep, nil
}

func (a *EnterpriseAdapter) diffWSDL(base, head []byte) (*report.DiffReport, error) {
	var baseDef, headDef Definitions

	if err := xml.Unmarshal(base, &baseDef); err != nil {
		return nil, fmt.Errorf("failed to parse base WSDL: %w", err)
	}

	if err := xml.Unmarshal(head, &headDef); err != nil {
		return nil, fmt.Errorf("failed to parse head WSDL: %w", err)
	}

	rep := &report.DiffReport{
		SubstrateVersion: "0.1.0",
		SchemaType:       "wsdl",
		ComparedAt:       time.Now().UTC().Format(time.RFC3339),
		Summary: report.Summary{
			OverallSeverity: report.SeverityNoChanges,
		},
		BreakingChanges: []report.Change{},
		Warnings:        []report.Change{},
		SafeChanges:     []report.Change{},
	}

	baseOps := make(map[string]map[string]Operation) // portType -> operationName -> Operation
	for _, pt := range baseDef.PortType {
		if baseOps[pt.Name] == nil {
			baseOps[pt.Name] = make(map[string]Operation)
		}
		for _, op := range pt.Operation {
			baseOps[pt.Name][op.Name] = op
		}
	}

	headOps := make(map[string]map[string]Operation)
	for _, pt := range headDef.PortType {
		if headOps[pt.Name] == nil {
			headOps[pt.Name] = make(map[string]Operation)
		}
		for _, op := range pt.Operation {
			headOps[pt.Name][op.Name] = op
		}
	}

	for ptName, bOps := range baseOps {
		hOps, ptExists := headOps[ptName]
		if !ptExists {
			// If PortType is removed, all operations inside it are removed
			for opName := range bOps {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:          fmt.Sprintf("wsdl-operation-removed-%s-%s", ptName, opName),
					RuleID:      "WSDL_OPERATION_REMOVED",
					Severity:    report.ChangeSeverityBreaking,
					Path:        fmt.Sprintf("portType.%s.operation.%s", ptName, opName),
					Description: fmt.Sprintf("WSDL operation '%s' removed from portType '%s'", opName, ptName),
					Before:      opName,
					After:       nil,
				})
			}
			continue
		}

		for opName, bOp := range bOps {
			hOp, opExists := hOps[opName]
			if !opExists {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:          fmt.Sprintf("wsdl-operation-removed-%s-%s", ptName, opName),
					RuleID:      "WSDL_OPERATION_REMOVED",
					Severity:    report.ChangeSeverityBreaking,
					Path:        fmt.Sprintf("portType.%s.operation.%s", ptName, opName),
					Description: fmt.Sprintf("WSDL operation '%s' removed from portType '%s'", opName, ptName),
					Before:      opName,
					After:       nil,
				})
				continue
			}

			if bOp.Input != nil && hOp.Input != nil {
				if bOp.Input.Message != hOp.Input.Message {
					rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
						ID:          fmt.Sprintf("wsdl-input-message-changed-%s-%s", ptName, opName),
						RuleID:      "WSDL_INPUT_MESSAGE_CHANGED",
						Severity:    report.ChangeSeverityBreaking,
						Path:        fmt.Sprintf("portType.%s.operation.%s.input.message", ptName, opName),
						Description: fmt.Sprintf("Input message for operation '%s' changed from '%s' to '%s'", opName, bOp.Input.Message, hOp.Input.Message),
						Before:      bOp.Input.Message,
						After:       hOp.Input.Message,
					})
				}
			} else if bOp.Input != nil && hOp.Input == nil {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:          fmt.Sprintf("wsdl-input-message-changed-%s-%s", ptName, opName),
					RuleID:      "WSDL_INPUT_MESSAGE_CHANGED",
					Severity:    report.ChangeSeverityBreaking,
					Path:        fmt.Sprintf("portType.%s.operation.%s.input", ptName, opName),
					Description: fmt.Sprintf("Input message for operation '%s' was removed", opName),
					Before:      bOp.Input.Message,
					After:       nil,
				})
			}

			baseFaults := make(map[string]Fault)
			for _, f := range bOp.Fault {
				baseFaults[f.Name] = f
			}

			headFaults := make(map[string]Fault)
			for _, f := range hOp.Fault {
				headFaults[f.Name] = f
			}

			for fName := range headFaults {
				if _, ok := baseFaults[fName]; !ok {
					rep.Warnings = append(rep.Warnings, report.Change{
						ID:          fmt.Sprintf("wsdl-fault-added-%s-%s-%s", ptName, opName, fName),
						RuleID:      "WSDL_FAULT_ADDED",
						Severity:    report.ChangeSeverityWarning,
						Path:        fmt.Sprintf("portType.%s.operation.%s.fault.%s", ptName, opName, fName),
						Description: fmt.Sprintf("Fault '%s' added to operation '%s'", fName, opName),
						Before:      nil,
						After:       fName,
					})
				}
			}
		}
	}

	baseBindings := make(map[string]Binding)
	for _, b := range baseDef.Binding {
		baseBindings[b.Name] = b
	}

	headBindings := make(map[string]Binding)
	for _, b := range headDef.Binding {
		headBindings[b.Name] = b
	}

	for bName := range baseBindings {
		if _, ok := headBindings[bName]; !ok {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("wsdl-binding-removed-%s", bName),
				RuleID:      "WSDL_BINDING_REMOVED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("binding.%s", bName),
				Description: fmt.Sprintf("WSDL binding '%s' was removed", bName),
				Before:      bName,
				After:       nil,
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

	return rep, nil
}
