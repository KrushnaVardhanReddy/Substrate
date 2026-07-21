package crm

import (
	"context"
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v78"
)

type mockStripeAPI struct {
	cust *stripe.Customer
	err  error
}

func (m *mockStripeAPI) GetCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	return m.cust, m.err
}

func TestStripeClient(t *testing.T) {
	client := NewStripeClient("dummy")
	assert.NotNil(t, client)
}

func TestStripeClient_GetCustomerMRR(t *testing.T) {
	client := NewStripeClient("dummy")

	// Test error
	client.api = &mockStripeAPI{err: errors.New("api error")}
	_, err := client.GetCustomerMRR(context.Background(), "cus_123")
	assert.Error(t, err)

	// Test success with no subs
	client.api = &mockStripeAPI{cust: &stripe.Customer{}}
	mrr, err := client.GetCustomerMRR(context.Background(), "cus_123")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, mrr)

	// Test success with subs
	client.api = &mockStripeAPI{
		cust: &stripe.Customer{
			Subscriptions: &stripe.SubscriptionList{
				Data: []*stripe.Subscription{
					{
						Status: stripe.SubscriptionStatusActive,
						Items: &stripe.SubscriptionItemList{
							Data: []*stripe.SubscriptionItem{
								{
									Price:    &stripe.Price{UnitAmount: 1000},
									Quantity: 2,
								},
							},
						},
					},
					{
						Status: stripe.SubscriptionStatusCanceled,
						Items: &stripe.SubscriptionItemList{
							Data: []*stripe.SubscriptionItem{
								{
									Price:    &stripe.Price{UnitAmount: 500},
									Quantity: 1,
								},
							},
						},
					},
				},
			},
		},
	}
	mrr, err = client.GetCustomerMRR(context.Background(), "cus_123")
	assert.NoError(t, err)
	assert.Equal(t, 20.0, mrr) // 1000/100 * 2 = 20
}
