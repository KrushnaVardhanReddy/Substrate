package provider

import (
	"context"

	"github.com/KrushnaVardhanReddy/substrate/terraform-provider-substrate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &webhookResource{}

func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	client *client.Client
}

type webhookResourceModel struct {
	Org    types.String `tfsdk:"org"`
	URL    types.String `tfsdk:"url"`
	Secret types.String `tfsdk:"secret"`
	ID     types.String `tfsdk:"id"`
}

func (r *webhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				Required:    true,
				Description: "Organization name",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "Webhook URL",
			},
			"secret": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Webhook secret",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Webhook ID",
			},
		},
	}
}

func (r *webhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *client.Client, got: %T. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = client
}

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhookReq := client.WebhookRequest{
		URL:    plan.URL.ValueString(),
		Secret: plan.Secret.ValueString(),
	}

	err := r.client.CreateWebhook(ctx, plan.Org.ValueString(), webhookReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating webhook", err.Error())
		return
	}

	// Since there is no actual GET endpoint returning an ID for created webhooks right now, we use a placeholder composite ID
	plan.ID = types.StringValue(plan.Org.ValueString() + "_" + plan.URL.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// The Substrate API currently does not support GET for a single webhook.
	// The creation endpoint acts as a register/update.
	// For Terraform, we'll assume it exists if the ID is set.
	var state webhookResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhookReq := client.WebhookRequest{
		URL:    plan.URL.ValueString(),
		Secret: plan.Secret.ValueString(),
	}

	err := r.client.CreateWebhook(ctx, plan.Org.ValueString(), webhookReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating webhook", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Org.ValueString() + "_" + plan.URL.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// API currently does not support DELETE
}
