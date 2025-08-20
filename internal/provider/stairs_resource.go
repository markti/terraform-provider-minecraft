package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = stairsResourceType{}
var _ tfsdk.Resource = stairsResource{}
var _ tfsdk.ResourceWithImportState = stairsResource{}

type stairsResourceType struct{}

func (t stairsResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A Minecraft stairs block (e.g., minecraft:oak_stairs) with orientation and shape.",
		Attributes: map[string]tfsdk.Attribute{
			"material": {
				MarkdownDescription: "The stairs material (e.g., `minecraft:oak_stairs`, `minecraft:stone_brick_stairs`).",
				Required:            true,
				Type:                types.StringType,
			},
			"position": {
				MarkdownDescription: "The position of the stairs block.",
				Required:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"x": {
						MarkdownDescription: "X coordinate of the block",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"y": {
						MarkdownDescription: "Y coordinate of the block",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"z": {
						MarkdownDescription: "Z coordinate of the block",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
				}),
			},

			// Stairs block states
			"facing": {
				MarkdownDescription: "Direction the stairs face: one of `north`, `south`, `east`, `west`.",
				Required:            true,
				Type:                types.StringType,
			},
			"half": {
				MarkdownDescription: "Whether the stairs are on the `top` (upside-down) or `bottom` half.",
				Required:            true,
				Type:                types.StringType,
			},
			"shape": {
				MarkdownDescription: "Stair shape: `straight`, `inner_left`, `inner_right`, `outer_left`, or `outer_right`.",
				Required:            true,
				Type:                types.StringType,
			},
			"waterlogged": {
				MarkdownDescription: "Whether the stairs are waterlogged.",
				Optional:            true,
				Type:                types.BoolType,
			},

			"id": {
				Computed:            true,
				MarkdownDescription: "ID of the block",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t stairsResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return stairsResource{provider: provider}, diags
}

type stairsResourceData struct {
	Id       types.String `tfsdk:"id"`
	Material string       `tfsdk:"material"`
	Position struct {
		X int `tfsdk:"x"`
		Y int `tfsdk:"y"`
		Z int `tfsdk:"z"`
	} `tfsdk:"position"`

	Facing      string `tfsdk:"facing"`      // north|south|east|west
	Half        string `tfsdk:"half"`        // top|bottom
	Shape       string `tfsdk:"shape"`       // straight|inner_left|inner_right|outer_left|outer_right
	Waterlogged *bool  `tfsdk:"waterlogged"` // optional
}

type stairsResource struct {
	provider provider
}

func (r stairsResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data stairsResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client, got error: %s", err))
		return
	}

	water := false
	if data.Waterlogged != nil {
		water = *data.Waterlogged
	}

	// Optional: guard materials if you want
	// if !strings.HasSuffix(data.Material, "_stairs") {
	// 	resp.Diagnostics.AddError("Validation Error", "material must be a *_stairs block")
	// 	return
	// }

	err = client.CreateStairs(
		ctx,
		data.Material,
		data.Position.X, data.Position.Y, data.Position.Z,
		// pass through as-is; server expects valid values
		data.Facing,
		data.Half,
		data.Shape,
		water,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create stairs, got error: %s", err))
		return
	}

	data.Id = types.String{Value: fmt.Sprintf("stairs-%d-%d-%d", data.Position.X, data.Position.Y, data.Position.Z)}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Read is a no-op; we don’t query Minecraft state (no stable read API).
func (r stairsResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data stairsResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r stairsResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data stairsResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client, got error: %s", err))
		return
	}

	water := false
	if data.Waterlogged != nil {
		water = *data.Waterlogged
	}

	err = client.CreateStairs(
		ctx,
		data.Material,
		data.Position.X, data.Position.Y, data.Position.Z,
		data.Facing,
		data.Half,
		data.Shape,
		water,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update stairs, got error: %s", err))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r stairsResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data stairsResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client, got error: %s", err))
		return
	}

	// Replace with air
	err = client.DeleteBlock(ctx, data.Position.X, data.Position.Y, data.Position.Z)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete block, got error: %s", err))
		return
	}
}

func (r stairsResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
