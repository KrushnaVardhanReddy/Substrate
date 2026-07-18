package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/terraform-provider-substrate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPolicyResource(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/org/test-org/enforce", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req client.PolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	providerConfig := fmt.Sprintf(`
provider "substrate" {
	base_url  = "%s"
	api_token = "test-token"
}
`, server.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"substrate": providerserver.NewProtocol6WithError(New("1.0.0")()),
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "substrate_policy" "test" {
  org     = "test-org"
  enforce = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("substrate_policy.test", "org", "test-org"),
					resource.TestCheckResourceAttr("substrate_policy.test", "enforce", "true"),
					resource.TestCheckResourceAttrSet("substrate_policy.test", "id"),
				),
			},
			// Update testing
			{
				Config: providerConfig + `
resource "substrate_policy" "test" {
  org     = "test-org"
  enforce = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("substrate_policy.test", "org", "test-org"),
					resource.TestCheckResourceAttr("substrate_policy.test", "enforce", "false"),
					resource.TestCheckResourceAttrSet("substrate_policy.test", "id"),
				),
			},
		},
	})
}
