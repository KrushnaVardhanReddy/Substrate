package provider

import (
	"context"

	"github.com/KrushnaVardhanReddy/substrate/terraform-provider-substrate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &substrateProvider{}

type substrateProvider struct {
	version string
}

type substrateProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	APIToken types.String `tfsdk:"api_token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &substrateProvider{
			version: version,
		}
	}
}

func (p *substrateProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "substrate"
	resp.Version = p.version
}

func (p *substrateProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "The Substrate API base URL (defaults to https://api.substrate.dev)",
			},
			"api_token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The Substrate API token",
			},
		},
	}
}

func (p *substrateProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config substrateProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := "https://api.substrate.dev"
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}

	apiToken := config.APIToken.ValueString()
	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider cannot create the Substrate API client as there is a missing or empty value for the Substrate API token.",
		)
		return
	}

	c := client.New(baseURL, apiToken)

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *substrateProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWebhookResource,
		NewPolicyResource,
	}
}

func (p *substrateProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
