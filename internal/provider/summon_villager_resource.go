package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = summonVillagerResourceType{}
var _ tfsdk.Resource = summonVillagerResource{}
var _ tfsdk.ResourceWithImportState = summonVillagerResource{}

type summonVillagerResourceType struct{}

func (t summonVillagerResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A Minecraft villager entity, summoned with optional NBT data tags and tracked by a stable UUID.",

		Attributes: map[string]tfsdk.Attribute{
			"x": {
				MarkdownDescription: "X coordinate where to summon the villager.",
				Required:            true,
				Type:                types.NumberType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // position can't change in-place
				},
			},
			"y": {
				MarkdownDescription: "Y coordinate where to summon the villager.",
				Required:            true,
				Type:                types.NumberType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // position can't change in-place
				},
			},
			"z": {
				MarkdownDescription: "Z coordinate where to summon the villager.",
				Required:            true,
				Type:                types.NumberType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // position can't change in-place
				},
			},
			"data_tag": {
				MarkdownDescription: "Optional NBT data tags for the villager as a JSON string. Example: `\"{\\\"VillagerData\\\": {\\\"profession\\\": \\\"farmer\\\", \\\"level\\\": 2, \\\"type\\\": \\\"plains\\\"}}\"` or `\"{\\\"Profession\\\": 1, \\\"Career\\\": 2, \\\"CareerLevel\\\": 3}\"`.",
				Optional:            true,
				Type:                types.StringType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "UUID for this villager (embedded as the entity's CustomName).",
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (t summonVillagerResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return summonVillagerResource{provider: provider}, diags
}

type summonVillagerResourceData struct {
	Id      types.String `tfsdk:"id"`
	X       int          `tfsdk:"x"`
	Y       int          `tfsdk:"y"`
	Z       int          `tfsdk:"z"`
	DataTag types.String `tfsdk:"data_tag"`
}

type summonVillagerResource struct {
	provider provider
}

func (r summonVillagerResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data summonVillagerResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	// Generate a stable UUID and use it as both TF id and the entity's CustomName.
	id := uuid.NewString()

	// Get the data tag as a string (JSON format)
	var dataTagJSON string
	if !data.DataTag.Null && !data.DataTag.Unknown {
		dataTagJSON = data.DataTag.Value
	}

	if err := client.SummonVillager(ctx, data.X, data.Y, data.Z, id, dataTagJSON); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to summon villager: %s", err))
		return
	}

	data.Id = types.String{Value: id}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r summonVillagerResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data summonVillagerResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Implement drift detection via a client method that searches for the villager by CustomName.
	// For now, keep state unchanged.
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r summonVillagerResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// All mutable fields are ForceNew; there's nothing to update in place.
	var data summonVillagerResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r summonVillagerResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data summonVillagerResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	if err := client.DeleteVillager(ctx, data.Id.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete villager: %s", err))
		return
	}
}

func (r summonVillagerResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Import by UUID (id). Caller supplies matching config (coordinates/data_tag) in HCL.
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}