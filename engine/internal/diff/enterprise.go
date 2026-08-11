// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package diff

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

type CustomObject struct {
	XMLName xml.Name      `xml:"CustomObject"`
	Fields  []CustomField `xml:"fields"`
}

type CustomField struct {
	FullName string `xml:"fullName"`
	Type     string `xml:"type"`
	Required bool   `xml:"required"`
	Length   int    `xml:"length"`
}

type Definitions struct {
	XMLName  xml.Name   `xml:"definitions"`
	PortType []PortType `xml:"portType"`
	Binding  []Binding  `xml:"binding"`
}

type PortType struct {
	Name      string      `xml:"name,attr"`
	Operation []Operation `xml:"operation"`
}

type Operation struct {
	Name   string  `xml:"name,attr"`
	Input  *Input  `xml:"input"`
	Output *Output `xml:"output"`
	Fault  []Fault `xml:"fault"`
}

type Input struct {
	Message string `xml:"message,attr"`
}

type Output struct {
	Message string `xml:"message,attr"`
}

type Fault struct {
	Name    string `xml:"name,attr"`
	Message string `xml:"message,attr"`
}

type Binding struct {
	Name string `xml:"name,attr"`
}

type EnterpriseAdapter struct{}

func (a *EnterpriseAdapter) Diff(base, head []byte, config map[string]interface{}) (*report.DiffReport, error) {
	if bytes.Contains(base, []byte("<CustomObject")) || bytes.Contains(head, []byte("<CustomObject")) {
		return a.diffSalesforce(base, head)
	}

	if bytes.Contains(base, []byte("<definitions")) || bytes.Contains(head, []byte("<definitions")) || bytes.Contains(base, []byte(":definitions")) || bytes.Contains(head, []byte(":definitions")) {
		return a.diffWSDL(base, head)
	}

	return nil, fmt.Errorf("unrecognized enterprise metadata format")
}
