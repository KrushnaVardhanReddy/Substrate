// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package graphql

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/compliance"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

func injectFederation(schema string) string {
	if !strings.Contains(schema, "directive @key") {
		schema += `
scalar _FieldSet
directive @key(fields: _FieldSet!, resolvable: Boolean = true) repeatable on OBJECT | INTERFACE
directive @requires(fields: _FieldSet!) on FIELD_DEFINITION
directive @provides(fields: _FieldSet!) on FIELD_DEFINITION
directive @external on FIELD_DEFINITION | OBJECT
directive @shareable on FIELD_DEFINITION | OBJECT
directive @link(url: String!, as: String, for: String, import: [String]) repeatable on SCHEMA
directive @override(from: String!) on FIELD_DEFINITION
directive @inaccessible on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ARGUMENT_DEFINITION | SCALAR | ENUM | ENUM_VALUE | INPUT_OBJECT | INPUT_FIELD_DEFINITION
directive @tag(name: String!) repeatable on FIELD_DEFINITION | INTERFACE | OBJECT | UNION | ARGUMENT_DEFINITION | SCALAR | ENUM | ENUM_VALUE | INPUT_OBJECT | INPUT_FIELD_DEFINITION
`
	}
	return schema
}

// CompareGraphQL compares two GraphQL schemas and returns a DiffReport.
func CompareGraphQL(basePath, headPath string) (*report.DiffReport, error) {
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base schema: %w", err)
	}
	headData, err := os.ReadFile(headPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read head schema: %w", err)
	}

	baseSchemaAST, err := parser.ParseSchema(&ast.Source{Input: injectFederation(string(baseData))})
	if err != nil {
		return nil, err
	}
	headSchemaAST, err := parser.ParseSchema(&ast.Source{Input: injectFederation(string(headData))})
	if err != nil {
		return nil, err
	}

	rep := &report.DiffReport{
		SchemaType:       report.SchemaTypeGraphQL,
		ComparedAt:       time.Now().UTC().Format(time.RFC3339),
		SubstrateVersion: "v0.1.0",
		BreakingChanges:  []report.Change{},
		Warnings:         []report.Change{},
		SafeChanges:      []report.Change{},
	}

	// 1. Diff Types
	for _, baseType := range baseSchemaAST.Definitions {
		headType := headSchemaAST.Definitions.ForName(baseType.Name)
		if headType == nil {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("gql-type-removed-%s", baseType.Name),
				RuleID:      "GQL_TYPE_REMOVED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        baseType.Name,
				Description: fmt.Sprintf("Type '%s' was removed.", baseType.Name),
				Before:      baseType.Name,
				After:       nil,
			})
			continue
		}

		// Diff based on kind
		if baseType.Kind == ast.Enum {
			diffEnum(rep, baseType, headType)
		} else if baseType.Kind == ast.Union {
			diffUnion(rep, baseType, headType)
		} else if baseType.Kind == ast.Object || baseType.Kind == ast.Interface || baseType.Kind == ast.InputObject {
			diffFields(rep, baseType, headType)
		}

		// Diff implemented interfaces for Objects
		if baseType.Kind == ast.Object {
			for _, baseInterface := range baseType.Interfaces {
				found := false
				for _, headInterface := range headType.Interfaces {
					if baseInterface == headInterface {
						found = true
						break
					}
				}
				if !found {
					rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
						ID:          fmt.Sprintf("gql-interface-removed-%s-%s", baseType.Name, baseInterface),
						RuleID:      "GQL_INTERFACE_REMOVED",
						Severity:    report.ChangeSeverityBreaking,
						Path:        fmt.Sprintf("%s.interfaces", baseType.Name),
						Description: fmt.Sprintf("Object type '%s' no longer implements interface '%s'.", baseType.Name, baseInterface),
						Before:      baseInterface,
						After:       nil,
					})
				}
			}
		}
	}

	// 2. Diff Directives
	for _, baseDir := range baseSchemaAST.Directives {
		name := baseDir.Name
		headDir := headSchemaAST.Directives.ForName(name)
		if headDir == nil {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("gql-directive-removed-%s", name),
				RuleID:      "GQL_DIRECTIVE_REMOVED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("@%s", name),
				Description: fmt.Sprintf("Directive '@%s' was removed.", name),
				Before:      name,
				After:       nil,
			})
			continue
		}

		// Check locations
		for _, baseLoc := range baseDir.Locations {
			found := false
			for _, headLoc := range headDir.Locations {
				if baseLoc == headLoc {
					found = true
					break
				}
			}
			if !found {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:          fmt.Sprintf("gql-directive-location-removed-%s-%s", name, baseLoc),
					RuleID:      "GQL_DIRECTIVE_LOCATION_REMOVED",
					Severity:    report.ChangeSeverityBreaking,
					Path:        fmt.Sprintf("@%s.locations", name),
					Description: fmt.Sprintf("Location '%s' was removed from directive '@%s'.", baseLoc, name),
					Before:      string(baseLoc),
					After:       nil,
				})
			}
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
	} else if rep.Summary.TotalChanges > 0 {
		rep.Summary.OverallSeverity = report.SeveritySafe
	} else {
		rep.Summary.OverallSeverity = report.SeverityNoChanges
	}

	compliance.Audit(rep)
	return rep, nil
}

func diffEnum(rep *report.DiffReport, baseType, headType *ast.Definition) {
	for _, baseVal := range baseType.EnumValues {
		headVal := headType.EnumValues.ForName(baseVal.Name)
		if headVal == nil {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("gql-enum-val-removed-%s-%s", baseType.Name, baseVal.Name),
				RuleID:      "GQL_ENUM_VALUE_REMOVED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("%s.%s", baseType.Name, baseVal.Name),
				Description: fmt.Sprintf("Enum value '%s' was removed from enum '%s'.", baseVal.Name, baseType.Name),
				Before:      baseVal.Name,
				After:       nil,
			})
		} else {
			baseDep := baseVal.Directives.ForName("deprecated")
			headDep := headVal.Directives.ForName("deprecated")
			if baseDep == nil && headDep != nil {
				rep.Warnings = append(rep.Warnings, report.Change{
					ID:          fmt.Sprintf("gql-enum-val-deprecated-%s-%s", baseType.Name, baseVal.Name),
					RuleID:      "GQL_ENUM_VALUE_DEPRECATED",
					Severity:    report.ChangeSeverityWarning,
					Path:        fmt.Sprintf("%s.%s", baseType.Name, baseVal.Name),
					Description: fmt.Sprintf("Enum value '%s' in enum '%s' was deprecated.", baseVal.Name, baseType.Name),
					Before:      false,
					After:       true,
				})
			}
		}
	}
}

func diffUnion(rep *report.DiffReport, baseType, headType *ast.Definition) {
	for _, baseMember := range baseType.Types {
		found := false
		for _, headMember := range headType.Types {
			if baseMember == headMember {
				found = true
				break
			}
		}
		if !found {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          fmt.Sprintf("gql-union-member-removed-%s-%s", baseType.Name, baseMember),
				RuleID:      "GQL_UNION_MEMBER_REMOVED",
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("%s.%s", baseType.Name, baseMember),
				Description: fmt.Sprintf("Member type '%s' was removed from union '%s'.", baseMember, baseType.Name),
				Before:      baseMember,
				After:       nil,
			})
		}
	}
}

func isKeyField(baseType *ast.Definition, fieldName string) bool {
	for _, dir := range baseType.Directives {
		if dir.Name == "key" {
			fieldsArg := dir.Arguments.ForName("fields")
			if fieldsArg != nil && fieldsArg.Value != nil {
				keyFields := strings.Fields(fieldsArg.Value.Raw)
				for _, kf := range keyFields {
					if kf == fieldName {
						return true
					}
				}
			}
		}
	}
	return false
}

func diffFields(rep *report.DiffReport, baseType, headType *ast.Definition) {
	for _, baseField := range baseType.Fields {
		headField := headType.Fields.ForName(baseField.Name)
		if headField == nil {
			isKey := isKeyField(baseType, baseField.Name)
			ruleID := "GQL_FIELD_REMOVED"
			id := fmt.Sprintf("gql-field-removed-%s-%s", baseType.Name, baseField.Name)
			if isKey {
				ruleID = "FederationKeyBroken"
				id = fmt.Sprintf("gql-federation-key-broken-%s-%s", baseType.Name, baseField.Name)
			}
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          id,
				RuleID:      ruleID,
				Severity:    report.ChangeSeverityBreaking,
				Path:        fmt.Sprintf("%s.%s", baseType.Name, baseField.Name),
				Description: fmt.Sprintf("Field '%s' was removed from '%s'.", baseField.Name, baseType.Name),
				Before:      baseField.Name,
				After:       nil,
			})
			continue
		}

		// Field type changed
		baseTypeStr := baseField.Type.String()
		headTypeStr := headField.Type.String()
		if baseTypeStr != headTypeStr {
			if baseTypeStr+"!" != headTypeStr {
				isKey := isKeyField(baseType, baseField.Name)
				ruleID := "GQL_FIELD_TYPE_CHANGED"
				id := fmt.Sprintf("gql-field-type-changed-%s-%s", baseType.Name, baseField.Name)
				if isKey {
					ruleID = "FederationKeyBroken"
					id = fmt.Sprintf("gql-federation-key-broken-%s-%s", baseType.Name, baseField.Name)
				}
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:          id,
					RuleID:      ruleID,
					Severity:    report.ChangeSeverityBreaking,
					Path:        fmt.Sprintf("%s.%s", baseType.Name, baseField.Name),
					Description: fmt.Sprintf("Return type of field '%s' changed from '%s' to '%s'.", baseField.Name, baseTypeStr, headTypeStr),
					Before:      baseTypeStr,
					After:       headTypeStr,
				})
			}
		}

		// Field deprecation
		baseDep := baseField.Directives.ForName("deprecated")
		headDep := headField.Directives.ForName("deprecated")
		if baseDep == nil && headDep != nil {
			rep.Warnings = append(rep.Warnings, report.Change{
				ID:          fmt.Sprintf("gql-field-deprecated-%s-%s", baseType.Name, baseField.Name),
				RuleID:      "GQL_FIELD_DEPRECATED",
				Severity:    report.ChangeSeverityWarning,
				Path:        fmt.Sprintf("%s.%s", baseType.Name, baseField.Name),
				Description: fmt.Sprintf("Field '%s' in '%s' was deprecated.", baseField.Name, baseType.Name),
				Before:      false,
				After:       true,
			})
		}

		// Diff Arguments (for Object/Interface fields)
		if baseType.Kind == ast.Object || baseType.Kind == ast.Interface {
			for _, baseArg := range baseField.Arguments {
				headArg := headField.Arguments.ForName(baseArg.Name)
				if headArg == nil {
					rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
						ID:          fmt.Sprintf("gql-arg-removed-%s-%s-%s", baseType.Name, baseField.Name, baseArg.Name),
						RuleID:      "GQL_ARGUMENT_REMOVED",
						Severity:    report.ChangeSeverityBreaking,
						Path:        fmt.Sprintf("%s.%s(%s)", baseType.Name, baseField.Name, baseArg.Name),
						Description: fmt.Sprintf("Argument '%s' was removed from field '%s'.", baseArg.Name, baseField.Name),
						Before:      baseArg.Name,
						After:       nil,
					})
				} else {
					if baseArg.Type.String() != headArg.Type.String() {
						rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
							ID:          fmt.Sprintf("gql-arg-type-changed-%s-%s-%s", baseType.Name, baseField.Name, baseArg.Name),
							RuleID:      "GQL_ARGUMENT_TYPE_CHANGED",
							Severity:    report.ChangeSeverityBreaking,
							Path:        fmt.Sprintf("%s.%s(%s)", baseType.Name, baseField.Name, baseArg.Name),
							Description: fmt.Sprintf("Type of argument '%s' changed from '%s' to '%s'.", baseArg.Name, baseArg.Type.String(), headArg.Type.String()),
							Before:      baseArg.Type.String(),
							After:       headArg.Type.String(),
						})
					}
				}
			}

			// Check for new required arguments
			for _, headArg := range headField.Arguments {
				if baseField.Arguments.ForName(headArg.Name) == nil {
					if headArg.Type.NonNull && headArg.DefaultValue == nil {
						rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
							ID:          fmt.Sprintf("gql-req-arg-added-%s-%s-%s", baseType.Name, baseField.Name, headArg.Name),
							RuleID:      "GQL_REQUIRED_ARGUMENT_ADDED",
							Severity:    report.ChangeSeverityBreaking,
							Path:        fmt.Sprintf("%s.%s(%s)", baseType.Name, baseField.Name, headArg.Name),
							Description: fmt.Sprintf("Required argument '%s' was added to field '%s'.", headArg.Name, baseField.Name),
							Before:      nil,
							After:       headArg.Name,
						})
					} else {
						rep.SafeChanges = append(rep.SafeChanges, report.Change{
							ID:          fmt.Sprintf("gql-opt-arg-added-%s-%s-%s", baseType.Name, baseField.Name, headArg.Name),
							RuleID:      "GQL_OPTIONAL_ARGUMENT_ADDED",
							Severity:    report.ChangeSeveritySafe,
							Path:        fmt.Sprintf("%s.%s(%s)", baseType.Name, baseField.Name, headArg.Name),
							Description: fmt.Sprintf("Optional argument '%s' was added to field '%s'.", headArg.Name, baseField.Name),
							Before:      nil,
							After:       headArg.Name,
						})
					}
				}
			}
		}
	}

	// Check for new input fields on InputObjects
	if headType.Kind == ast.InputObject {
		for _, headField := range headType.Fields {
			if baseType.Fields.ForName(headField.Name) == nil {
				if headField.Type.NonNull && headField.DefaultValue == nil {
					rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
						ID:          fmt.Sprintf("gql-req-input-field-added-%s-%s", headType.Name, headField.Name),
						RuleID:      "GQL_INPUT_FIELD_ADDED_REQUIRED",
						Severity:    report.ChangeSeverityBreaking,
						Path:        fmt.Sprintf("%s.%s", headType.Name, headField.Name),
						Description: fmt.Sprintf("Required input field '%s' was added to '%s'.", headField.Name, headType.Name),
						Before:      nil,
						After:       headField.Name,
					})
				} else {
					rep.SafeChanges = append(rep.SafeChanges, report.Change{
						ID:          fmt.Sprintf("gql-opt-input-field-added-%s-%s", headType.Name, headField.Name),
						RuleID:      "GQL_INPUT_FIELD_ADDED_OPTIONAL",
						Severity:    report.ChangeSeveritySafe,
						Path:        fmt.Sprintf("%s.%s", headType.Name, headField.Name),
						Description: fmt.Sprintf("Optional input field '%s' was added to '%s'.", headField.Name, headType.Name),
						Before:      nil,
						After:       headField.Name,
					})
				}
			}
		}
	} else if headType.Kind == ast.Object || headType.Kind == ast.Interface {
		for _, headField := range headType.Fields {
			if baseType.Fields.ForName(headField.Name) == nil {
				rep.SafeChanges = append(rep.SafeChanges, report.Change{
					ID:          fmt.Sprintf("gql-field-added-%s-%s", headType.Name, headField.Name),
					RuleID:      "GQL_FIELD_ADDED",
					Severity:    report.ChangeSeveritySafe,
					Path:        fmt.Sprintf("%s.%s", headType.Name, headField.Name),
					Description: fmt.Sprintf("Field '%s' was added to '%s'.", headField.Name, headType.Name),
					Before:      nil,
					After:       headField.Name,
				})
			}
		}
	}
}
