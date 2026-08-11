// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package diff

import (
	"testing"
)

func TestCompareEnterprise_Salesforce(t *testing.T) {
	adapter := &EnterpriseAdapter{}

	tests := []struct {
		name         string
		baseXML      string
		headXML      string
		wantErr      bool
		wantBreaking int
		wantWarning  int
		wantSafe     int
		wantRules    []string
	}{
		{
			name:         "SFDC_FIELD_REMOVED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>OldField</fullName><type>Text</type></fields></CustomObject>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><CustomObject></CustomObject>`,
			wantBreaking: 1,
			wantRules:    []string{"SFDC_FIELD_REMOVED"},
		},
		{
			name:         "SFDC_FIELD_REQUIRED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>MyField</fullName><type>Text</type><required>false</required></fields></CustomObject>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>MyField</fullName><type>Text</type><required>true</required></fields></CustomObject>`,
			wantBreaking: 1,
			wantRules:    []string{"SFDC_FIELD_REQUIRED"},
		},
		{
			name:         "SFDC_FIELD_TYPE_CHANGED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>MyField</fullName><type>Text</type></fields></CustomObject>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>MyField</fullName><type>Number</type></fields></CustomObject>`,
			wantBreaking: 1,
			wantRules:    []string{"SFDC_FIELD_TYPE_CHANGED"},
		},
		{
			name:         "SFDC_FIELD_LENGTH_REDUCED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>MyField</fullName><type>Text</type><length>255</length></fields></CustomObject>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>MyField</fullName><type>Text</type><length>100</length></fields></CustomObject>`,
			wantBreaking: 1,
			wantRules:    []string{"SFDC_FIELD_LENGTH_REDUCED"},
		},
		{
			name:      "SFDC_FIELD_ADDED",
			baseXML:   `<?xml version="1.0" encoding="UTF-8"?><CustomObject></CustomObject>`,
			headXML:   `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>NewField</fullName><type>Text</type></fields></CustomObject>`,
			wantSafe:  1,
			wantRules: []string{"SFDC_FIELD_ADDED"},
		},
		{
			name:    "Invalid XML Base",
			baseXML: `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>NewField</fullName>`,
			headXML: `<?xml version="1.0" encoding="UTF-8"?><CustomObject></CustomObject>`,
			wantErr: true,
		},
		{
			name:    "Invalid XML Head",
			baseXML: `<?xml version="1.0" encoding="UTF-8"?><CustomObject></CustomObject>`,
			headXML: `<?xml version="1.0" encoding="UTF-8"?><CustomObject><fields><fullName>NewField</fullName>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := adapter.Diff([]byte(tt.baseXML), []byte(tt.headXML), nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Diff() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if rep.Summary.BreakingCount != tt.wantBreaking {
				t.Errorf("Diff() BreakingCount = %v, want %v", rep.Summary.BreakingCount, tt.wantBreaking)
			}
			if rep.Summary.WarningCount != tt.wantWarning {
				t.Errorf("Diff() WarningCount = %v, want %v", rep.Summary.WarningCount, tt.wantWarning)
			}
			if rep.Summary.SafeCount != tt.wantSafe {
				t.Errorf("Diff() SafeCount = %v, want %v", rep.Summary.SafeCount, tt.wantSafe)
			}

			// verify rule IDs
			var foundRules []string
			for _, c := range rep.BreakingChanges {
				foundRules = append(foundRules, string(c.RuleID))
			}
			for _, c := range rep.Warnings {
				foundRules = append(foundRules, string(c.RuleID))
			}
			for _, c := range rep.SafeChanges {
				foundRules = append(foundRules, string(c.RuleID))
			}

			for _, wantRule := range tt.wantRules {
				found := false
				for _, fr := range foundRules {
					if fr == wantRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Diff() expected rule %s not found in changes: %v", wantRule, foundRules)
				}
			}
		})
	}
}

func TestCompareEnterprise_WSDL(t *testing.T) {
	adapter := &EnterpriseAdapter{}

	tests := []struct {
		name         string
		baseXML      string
		headXML      string
		wantErr      bool
		wantBreaking int
		wantWarning  int
		wantSafe     int
		wantRules    []string
	}{
		{
			name:         "WSDL_OPERATION_REMOVED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"><operation name="OldOp"></operation></portType></definitions>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"></portType></definitions>`,
			wantBreaking: 1,
			wantRules:    []string{"WSDL_OPERATION_REMOVED"},
		},
		{
			name:         "WSDL_PORTTYPE_REMOVED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"><operation name="OldOp"></operation></portType></definitions>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions></definitions>`,
			wantBreaking: 1,
			wantRules:    []string{"WSDL_OPERATION_REMOVED"},
		},
		{
			name:         "WSDL_INPUT_MESSAGE_CHANGED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"><operation name="Op"><input message="Msg1"/></operation></portType></definitions>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"><operation name="Op"><input message="Msg2"/></operation></portType></definitions>`,
			wantBreaking: 1,
			wantRules:    []string{"WSDL_INPUT_MESSAGE_CHANGED"},
		},
		{
			name:         "WSDL_INPUT_MESSAGE_REMOVED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"><operation name="Op"><input message="Msg1"/></operation></portType></definitions>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"><operation name="Op"></operation></portType></definitions>`,
			wantBreaking: 1,
			wantRules:    []string{"WSDL_INPUT_MESSAGE_CHANGED"},
		},
		{
			name:        "WSDL_FAULT_ADDED",
			baseXML:     `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"><operation name="Op"></operation></portType></definitions>`,
			headXML:     `<?xml version="1.0" encoding="UTF-8"?><definitions><portType name="MyPort"><operation name="Op"><fault name="Fault1" message="MsgFault"/></operation></portType></definitions>`,
			wantWarning: 1,
			wantRules:   []string{"WSDL_FAULT_ADDED"},
		},
		{
			name:         "WSDL_BINDING_REMOVED",
			baseXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions><binding name="MyBinding"></binding></definitions>`,
			headXML:      `<?xml version="1.0" encoding="UTF-8"?><definitions></definitions>`,
			wantBreaking: 1,
			wantRules:    []string{"WSDL_BINDING_REMOVED"},
		},
		{
			name:    "Invalid XML Base",
			baseXML: `<?xml version="1.0" encoding="UTF-8"?><definitions><binding name="MyBinding">`,
			headXML: `<?xml version="1.0" encoding="UTF-8"?><definitions></definitions>`,
			wantErr: true,
		},
		{
			name:    "Invalid XML Head",
			baseXML: `<?xml version="1.0" encoding="UTF-8"?><definitions></definitions>`,
			headXML: `<?xml version="1.0" encoding="UTF-8"?><definitions><binding name="MyBinding">`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := adapter.Diff([]byte(tt.baseXML), []byte(tt.headXML), nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Diff() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if rep.Summary.BreakingCount != tt.wantBreaking {
				t.Errorf("Diff() BreakingCount = %v, want %v", rep.Summary.BreakingCount, tt.wantBreaking)
			}
			if rep.Summary.WarningCount != tt.wantWarning {
				t.Errorf("Diff() WarningCount = %v, want %v", rep.Summary.WarningCount, tt.wantWarning)
			}
			if rep.Summary.SafeCount != tt.wantSafe {
				t.Errorf("Diff() SafeCount = %v, want %v", rep.Summary.SafeCount, tt.wantSafe)
			}

			// verify rule IDs
			var foundRules []string
			for _, c := range rep.BreakingChanges {
				foundRules = append(foundRules, string(c.RuleID))
			}
			for _, c := range rep.Warnings {
				foundRules = append(foundRules, string(c.RuleID))
			}
			for _, c := range rep.SafeChanges {
				foundRules = append(foundRules, string(c.RuleID))
			}

			for _, wantRule := range tt.wantRules {
				found := false
				for _, fr := range foundRules {
					if fr == wantRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Diff() expected rule %s not found in changes: %v", wantRule, foundRules)
				}
			}
		})
	}
}

func TestCompareEnterprise_Unrecognized(t *testing.T) {
	adapter := &EnterpriseAdapter{}
	_, err := adapter.Diff([]byte("random data"), []byte("random data"), nil)
	if err == nil {
		t.Fatal("expected error for unrecognized format")
	}
}
