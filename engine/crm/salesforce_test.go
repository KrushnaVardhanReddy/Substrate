// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package crm

import (
	"context"
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/simpleforce/simpleforce"
)

type mockSalesforceAPI struct {
	queryResult *simpleforce.QueryResult
	queryErr    error
	loginErr    error
}

func (m *mockSalesforceAPI) Query(query string) (*simpleforce.QueryResult, error) {
	return m.queryResult, m.queryErr
}

func (m *mockSalesforceAPI) LoginPassword(username, password, token string) error {
	return m.loginErr
}

func TestSalesforceClient_InitError(t *testing.T) {
	// Real client login error
	client, err := NewSalesforceClient("https://login.salesforce.com", "dummy", "dummy", "dummy", "dummy", "dummy")
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestSalesforceClient_GetCustomerRevenue(t *testing.T) {
	client := &SalesforceClient{api: &mockSalesforceAPI{queryErr: errors.New("api error")}}

	// Error test
	_, err := client.GetCustomerRevenue(context.Background(), "acc_123")
	assert.Error(t, err)

	// No records
	client.api = &mockSalesforceAPI{queryResult: &simpleforce.QueryResult{TotalSize: 0}}
	rev, err := client.GetCustomerRevenue(context.Background(), "acc_123")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, rev)

	// Valid record
	client.api = &mockSalesforceAPI{
		queryResult: &simpleforce.QueryResult{
			TotalSize: 1,
			Records: []simpleforce.SObject{
				{"AnnualRevenue": 150000.0},
			},
		},
	}
	rev, err = client.GetCustomerRevenue(context.Background(), "acc_123")
	assert.NoError(t, err)
	assert.Equal(t, 150000.0, rev)

	// Missing field
	client.api = &mockSalesforceAPI{
		queryResult: &simpleforce.QueryResult{
			TotalSize: 1,
			Records: []simpleforce.SObject{
				{"OtherField": 150000.0},
			},
		},
	}
	rev, err = client.GetCustomerRevenue(context.Background(), "acc_123")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, rev)
}
