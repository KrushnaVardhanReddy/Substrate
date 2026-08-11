// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package crm

import (
	"context"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/customer"
)

type StripeAPI interface {
	GetCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error)
}

type realStripeAPI struct{}

func (r *realStripeAPI) GetCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	return customer.Get(id, params)
}

type StripeClient struct {
	APIKey string
	api    StripeAPI
}

func NewStripeClient(apiKey string) *StripeClient {
	return &StripeClient{APIKey: apiKey, api: &realStripeAPI{}}
}

func (c *StripeClient) GetCustomerMRR(ctx context.Context, customerID string) (float64, error) {
	stripe.Key = c.APIKey

	params := &stripe.CustomerParams{}
	params.AddExpand("subscriptions")

	cust, err := c.api.GetCustomer(customerID, params)
	if err != nil {
		return 0, err
	}

	var mrr float64
	if cust.Subscriptions != nil && cust.Subscriptions.Data != nil {
		for _, sub := range cust.Subscriptions.Data {
			if sub.Status == stripe.SubscriptionStatusActive && sub.Items != nil {
				for _, item := range sub.Items.Data {
					if item.Price != nil && item.Price.UnitAmount > 0 {
						// Simple MRR calc for demo (assuming monthly)
						mrr += float64(item.Price.UnitAmount) / 100.0 * float64(item.Quantity)
					}
				}
			}
		}
	}

	return mrr, nil
}
