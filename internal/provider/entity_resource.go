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
var _ tfsdk.ResourceType = entityResourceType{}
var _ tfsdk.Resource = entityResource{}
var _ tfsdk.ResourceWithImportState = entityResource{}

type entityResourceType struct{}

func (t entityResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A Minecraft entity, summoned and tracked by a stable UUID.",

		Attributes: map[string]tfsdk.Attribute{
			"type": {
				MarkdownDescription: "The entity type (e.g. `minecraft:armor_stand`, `minecraft:text_display`).",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // entity kind can't change in-place
				},
			},
			"position": {
				MarkdownDescription: "The position to summon the entity at.",
				Required:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"x": {
						MarkdownDescription: "X coordinate",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"y": {
						MarkdownDescription: "Y coordinate",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"z": {
						MarkdownDescription: "Z coordinate",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
				}),
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "UUID for this entity (also embedded as the entity's CustomName/tag).",
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (t entityResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return entityResource{provider: provider}, diags
}

type entityResourceData struct {
	Id       types.String `tfsdk:"id"`
	Type     string       `tfsdk:"type"`
	Position struct {
		X int `tfsdk:"x"`
		Y int `tfsdk:"y"`
		Z int `tfsdk:"z"`
	} `tfsdk:"position"`
}

type entityResource struct {
	provider provider
}

func (r entityResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data entityResourceData
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

	// Generate a stable UUID and use it as both TF id and the entity's tag/CustomName.
	id := uuid.NewString()
	pos := fmt.Sprintf("%d %d %d", data.Position.X, data.Position.Y, data.Position.Z)

	if err := client.CreateEntity(ctx, data.Type, pos, id); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to summon entity: %s", err))
		return
	}

	data.Id = types.String{Value: id}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r entityResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data entityResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Implement drift detection via a client.GetEntity(ctx, type, id) that searches by tag/CustomName.
	// For now, keep state unchanged.
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r entityResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// All mutable fields are ForceNew; there's nothing to update in place.
	var data entityResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r entityResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data entityResourceData
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

	pos := fmt.Sprintf("%d %d %d", data.Position.X, data.Position.Y, data.Position.Z)
	if err := client.DeleteEntity(ctx, data.Type, pos, data.Id.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete entity: %s", err))
		return
	}
}

func (r entityResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Import by UUID (id). Caller supplies matching config (type/position) in HCL.
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
