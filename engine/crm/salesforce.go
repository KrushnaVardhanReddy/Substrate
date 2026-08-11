// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package crm

import (
	"context"
	"strings"

	"github.com/simpleforce/simpleforce"
)

type SalesforceAPI interface {
	Query(query string) (*simpleforce.QueryResult, error)
	LoginPassword(username, password, token string) error
}

type realSalesforceAPI struct {
	client *simpleforce.Client
}

func (r *realSalesforceAPI) Query(query string) (*simpleforce.QueryResult, error) {
	return r.client.Query(query)
}

func (r *realSalesforceAPI) LoginPassword(username, password, token string) error {
	return r.client.LoginPassword(username, password, token)
}

type SalesforceClient struct {
	api SalesforceAPI
}

func NewSalesforceClient(url, clientID, clientSecret, username, password, token string) (*SalesforceClient, error) {
	client := simpleforce.NewClient(url, clientID, "34.0")
	realAPI := &realSalesforceAPI{client: client}
	err := realAPI.LoginPassword(username, password, token)
	if err != nil {
		return nil, err
	}
	return &SalesforceClient{api: realAPI}, nil
}

func (c *SalesforceClient) GetCustomerRevenue(ctx context.Context, accountID string) (float64, error) {
	// Simple sanitization to prevent SOQL injection
	sanitizedAccountID := strings.ReplaceAll(accountID, "'", "\\'")
	query := "SELECT AnnualRevenue FROM Account WHERE Id = '" + sanitizedAccountID + "'"
	result, err := c.api.Query(query)
	if err != nil {
		return 0, err
	}

	if result.TotalSize > 0 {
		records := result.Records
		if len(records) > 0 {
			record := records[0]
			revenueVal := record["AnnualRevenue"]
			if revenue, ok := revenueVal.(float64); ok {
				return revenue, nil
			}
		}
	}
	return 0, nil
}
