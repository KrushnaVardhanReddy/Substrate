package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProvider_Validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"substrate": providerserver.NewProtocol6WithError(New("1.0.0")()),
		},
		Steps: []resource.TestStep{
			{
				Config: `
provider "substrate" {
}
resource "substrate_webhook" "test" {
  org    = "test-org"
  url    = "https://example.com"
  secret = "foo"
}
`,
				ExpectError: regexp.MustCompile(`The argument "api_token" is required`),
			},
		},
	})
}
