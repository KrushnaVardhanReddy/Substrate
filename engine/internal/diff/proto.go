// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package diff

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/compliance"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/emicklei/proto"
)

type ProtobufSchema struct {
	Messages map[string]*proto.Message
	Services map[string]*proto.Service
	Enums    map[string]*proto.Enum
}

func slugify(s string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(s, "_"))
}

func ParseProtobufDir(dirPath string) (*ProtobufSchema, error) {
	schema := &ProtobufSchema{
		Messages: make(map[string]*proto.Message),
		Services: make(map[string]*proto.Service),
		Enums:    make(map[string]*proto.Enum),
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", dirPath, err)
	}

	if !info.IsDir() {
		return parseProtobufFile(dirPath, schema)
	}

	err = filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
			_, err = parseProtobufFile(path, schema)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return schema, nil
}

func parseProtobufFile(path string, schema *ProtobufSchema) (*ProtobufSchema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	parser := proto.NewParser(f)
	definition, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto file %s: %w", path, err)
	}

	proto.Walk(definition,
		proto.WithService(func(s *proto.Service) {
			schema.Services[s.Name] = s
		}),
		proto.WithMessage(func(m *proto.Message) {
			schema.Messages[m.Name] = m
		}),
		proto.WithEnum(func(e *proto.Enum) {
			schema.Enums[e.Name] = e
		}),
	)

	return schema, nil
}

func CompareProto(baseDir, headDir string) (*report.DiffReport, error) {
	baseSchema, err := ParseProtobufDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base proto dir: %w", err)
	}

	headSchema, err := ParseProtobufDir(headDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse head proto dir: %w", err)
	}

	rep := &report.DiffReport{
		SubstrateVersion: "0.1.0",
		SchemaType:       report.SchemaTypeProtobuf,
		ComparedAt:       time.Now().UTC().Format(time.RFC3339),
		BreakingChanges:  []report.Change{},
		Warnings:         []report.Change{},
		SafeChanges:      []report.Change{},
	}

	// Compare Messages (Fields)
	for name, baseMsg := range baseSchema.Messages {
		headMsg, ok := headSchema.Messages[name]
		if !ok {
			// Message removed
			// Depending on usage, this might not be strictly breaking wire protocol if it's not referenced,
			// but for schema compatibility, it's breaking.
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:             fmt.Sprintf("proto_message_removed_%s", slugify(name)),
				RuleID:         "PROTO_MESSAGE_REMOVED",
				Severity:       report.ChangeSeverityBreaking,
				Path:           name,
				Description:    fmt.Sprintf("Message %s was removed", name),
				Recommendation: func(s string) *string { return &s }("Do not remove messages. Deprecate them instead."),
			})
			continue
		}

		baseFields := make(map[string]*proto.NormalField)
		baseFieldsBySeq := make(map[int]*proto.NormalField)
		for _, e := range baseMsg.Elements {
			if f, ok := e.(*proto.NormalField); ok {
				baseFields[f.Name] = f
				baseFieldsBySeq[f.Sequence] = f
			}
		}

		headFields := make(map[string]*proto.NormalField)
		for _, e := range headMsg.Elements {
			if f, ok := e.(*proto.NormalField); ok {
				headFields[f.Name] = f
			}
		}

		// Check for removed or changed fields
		for fName, baseField := range baseFields {
			headField, ok := headFields[fName]
			if !ok {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("proto_field_removed_%s_%s", slugify(name), slugify(fName)),
					RuleID:         "PROTO_FIELD_REMOVED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("%s.%s", name, fName),
					Description:    fmt.Sprintf("Field %s was removed from message %s", fName, name),
					Recommendation: func(s string) *string { return &s }("Do not remove fields. Mark as deprecated instead."),
				})
				continue
			}

			if baseField.Sequence != headField.Sequence {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("proto_field_number_changed_%s_%s", slugify(name), slugify(fName)),
					RuleID:         "PROTO_FIELD_NUMBER_CHANGED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("%s.%s", name, fName),
					Description:    fmt.Sprintf("Field %s in message %s changed number from %d to %d", fName, name, baseField.Sequence, headField.Sequence),
					Recommendation: func(s string) *string { return &s }("Field numbers are wire-protocol identifiers and must never change."),
				})
			}

			if baseField.Type != headField.Type {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("proto_field_type_changed_%s_%s", slugify(name), slugify(fName)),
					RuleID:         "PROTO_FIELD_TYPE_CHANGED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("%s.%s", name, fName),
					Description:    fmt.Sprintf("Field %s in message %s changed type from %s to %s", fName, name, baseField.Type, headField.Type),
					Recommendation: func(s string) *string { return &s }("Field type changes break wire compatibility. Add a new field instead."),
				})
			}
		}

		// Check for added fields (safe change)
		for fName, headField := range headFields {
			if _, ok := baseFields[fName]; !ok {
				// Added a new field. Check if sequence overlaps
				if _, seqConflict := baseFieldsBySeq[headField.Sequence]; seqConflict {
					// We might already flag this as a tag change if they renamed a field and kept the tag,
					// but let's be safe.
				}
				rep.SafeChanges = append(rep.SafeChanges, report.Change{
					ID:          fmt.Sprintf("proto_field_added_%s_%s", slugify(name), slugify(fName)),
					RuleID:      "PROTO_FIELD_ADDED",
					Severity:    report.ChangeSeveritySafe,
					Path:        fmt.Sprintf("%s.%s", name, fName),
					Description: fmt.Sprintf("Field %s was added to message %s", fName, name),
				})
			}
		}
	}

	// Compare Enums
	for name, baseEnum := range baseSchema.Enums {
		headEnum, ok := headSchema.Enums[name]
		if !ok {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:             fmt.Sprintf("proto_enum_removed_%s", slugify(name)),
				RuleID:         "PROTO_ENUM_REMOVED",
				Severity:       report.ChangeSeverityBreaking,
				Path:           name,
				Description:    fmt.Sprintf("Enum %s was removed", name),
				Recommendation: func(s string) *string { return &s }("Do not remove enums. Deprecate them instead."),
			})
			continue
		}

		baseFields := make(map[string]*proto.EnumField)
		for _, e := range baseEnum.Elements {
			if f, ok := e.(*proto.EnumField); ok {
				baseFields[f.Name] = f
			}
		}
		headFields := make(map[string]*proto.EnumField)
		for _, e := range headEnum.Elements {
			if f, ok := e.(*proto.EnumField); ok {
				headFields[f.Name] = f
			}
		}

		for fName := range baseFields {
			if _, ok := headFields[fName]; !ok {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("proto_enum_value_removed_%s_%s", slugify(name), slugify(fName)),
					RuleID:         "PROTO_ENUM_VALUE_REMOVED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("%s.%s", name, fName),
					Description:    fmt.Sprintf("Enum value %s was removed from %s", fName, name),
					Recommendation: func(s string) *string { return &s }("Removing enum values breaks clients that send or receive that value."),
				})
			}
		}
	}

	// Compare Services and RPCs
	for name, baseSvc := range baseSchema.Services {
		headSvc, ok := headSchema.Services[name]
		if !ok {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:             fmt.Sprintf("proto_service_removed_%s", slugify(name)),
				RuleID:         "PROTO_SERVICE_REMOVED",
				Severity:       report.ChangeSeverityBreaking,
				Path:           name,
				Description:    fmt.Sprintf("Service %s was removed", name),
				Recommendation: func(s string) *string { return &s }("Removing a service breaks all clients. Use a deprecation notice first."),
			})
			continue
		}

		baseRPCs := make(map[string]*proto.RPC)
		for _, e := range baseSvc.Elements {
			if rpc, ok := e.(*proto.RPC); ok {
				baseRPCs[rpc.Name] = rpc
			}
		}
		headRPCs := make(map[string]*proto.RPC)
		for _, e := range headSvc.Elements {
			if rpc, ok := e.(*proto.RPC); ok {
				headRPCs[rpc.Name] = rpc
			}
		}

		for rpcName, baseRPC := range baseRPCs {
			headRPC, ok := headRPCs[rpcName]
			if !ok {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("proto_rpc_removed_%s_%s", slugify(name), slugify(rpcName)),
					RuleID:         "PROTO_RPC_REMOVED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("%s.%s", name, rpcName),
					Description:    fmt.Sprintf("RPC method %s was removed from service %s", rpcName, name),
					Recommendation: func(s string) *string { return &s }("Removing an RPC method breaks all clients. Deprecate it first."),
				})
				continue
			}

			if baseRPC.RequestType != headRPC.RequestType {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("proto_rpc_req_type_changed_%s_%s", slugify(name), slugify(rpcName)),
					RuleID:         "PROTO_RPC_REQUEST_TYPE_CHANGED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("%s.%s", name, rpcName),
					Description:    fmt.Sprintf("RPC method %s request type changed from %s to %s", rpcName, baseRPC.RequestType, headRPC.RequestType),
					Recommendation: func(s string) *string { return &s }("Changing an RPC request type breaks clients."),
				})
			}

			if baseRPC.ReturnsType != headRPC.ReturnsType {
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("proto_rpc_resp_type_changed_%s_%s", slugify(name), slugify(rpcName)),
					RuleID:         "PROTO_RPC_RESPONSE_TYPE_CHANGED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("%s.%s", name, rpcName),
					Description:    fmt.Sprintf("RPC method %s response type changed from %s to %s", rpcName, baseRPC.ReturnsType, headRPC.ReturnsType),
					Recommendation: func(s string) *string { return &s }("Changing an RPC response type breaks clients."),
				})
			}
		}

		// Check for added RPCs (safe)
		for rpcName := range headRPCs {
			if _, ok := baseRPCs[rpcName]; !ok {
				rep.SafeChanges = append(rep.SafeChanges, report.Change{
					ID:          fmt.Sprintf("proto_rpc_added_%s_%s", slugify(name), slugify(rpcName)),
					RuleID:      "PROTO_RPC_ADDED",
					Severity:    report.ChangeSeveritySafe,
					Path:        fmt.Sprintf("%s.%s", name, rpcName),
					Description: fmt.Sprintf("RPC method %s was added to service %s", rpcName, name),
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
	} else if rep.Summary.SafeCount > 0 {
		rep.Summary.OverallSeverity = report.SeveritySafe
	} else {
		rep.Summary.OverallSeverity = report.SeverityNoChanges
	}

	compliance.Audit(rep)

	return rep, nil
}
