package provider

import (
	"context"

	"github.com/KrushnaVardhanReddy/substrate/terraform-provider-substrate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &policyResource{}

func NewPolicyResource() resource.Resource {
	return &policyResource{}
}

type policyResource struct {
	client *client.Client
}

type policyResourceModel struct {
	Org     types.String `tfsdk:"org"`
	Enforce types.Bool   `tfsdk:"enforce"`
	ID      types.String `tfsdk:"id"`
}

func (r *policyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				Required:    true,
				Description: "Organization name",
			},
			"enforce": schema.BoolAttribute{
				Required:    true,
				Description: "Whether policies are enforced globally",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Policy ID",
			},
		},
	}
}

func (r *policyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyReq := client.PolicyRequest{
		Enforce: plan.Enforce.ValueBool(),
	}

	err := r.client.SetPolicy(ctx, plan.Org.ValueString(), policyReq)
	if err != nil {
		resp.Diagnostics.AddError("Error setting policy", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Org.ValueString() + "_policy")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Substrate API currently doesn't have a GET for enforcement policies,
	// so we assume what is in state is accurate for now.
	var state policyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyReq := client.PolicyRequest{
		Enforce: plan.Enforce.ValueBool(),
	}

	err := r.client.SetPolicy(ctx, plan.Org.ValueString(), policyReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policy", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Org.ValueString() + "_policy")

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state policyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete can just be setting enforce to false
	policyReq := client.PolicyRequest{
		Enforce: false,
	}

	err := r.client.SetPolicy(ctx, state.Org.ValueString(), policyReq)
	if err != nil {
		resp.Diagnostics.AddError("Error removing policy", err.Error())
		return
	}
}
