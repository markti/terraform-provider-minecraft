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
var _ tfsdk.ResourceType = bedResourceType{}
var _ tfsdk.Resource = bedResource{}
var _ tfsdk.ResourceWithImportState = bedResource{}

type bedResourceType struct{}

func (t bedResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A Minecraft bed (two-block structure). The start position is the FOOT. Direction places the HEAD one block in that direction.",
		Attributes: map[string]tfsdk.Attribute{
			"material": {
				MarkdownDescription: "The bed material, e.g. `minecraft:red_bed`, `minecraft:blue_bed`.",
				Required:            true,
				Type:                types.StringType,
			},
			"position": {
				MarkdownDescription: "The FOOT position of the bed.",
				Required:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"x": {
						MarkdownDescription: "X coordinate (foot)",
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
						MarkdownDescription: "Z coordinate (foot)",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
				}),
			},
			"direction": {
				MarkdownDescription: "Direction the bed faces: one of `north`, `south`, `east`, `west`. The HEAD goes one block in this direction from the FOOT.",
				Required:            true,
				Type:                types.StringType,
			},
			// Optional convenience flag (defaults handled in code as false)
			"occupied": {
				MarkdownDescription: "Whether the bed is considered occupied (rarely needed). Defaults to false.",
				Optional:            true,
				Type:                types.BoolType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "ID of the bed resource.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t bedResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return bedResource{provider: provider}, diags
}

type bedResourceData struct {
	Id       types.String `tfsdk:"id"`
	Material string       `tfsdk:"material"`
	Position struct {
		X int `tfsdk:"x"`
		Y int `tfsdk:"y"`
		Z int `tfsdk:"z"`
	} `tfsdk:"position"`
	Direction string `tfsdk:"direction"` // north|south|east|west
	Occupied  *bool  `tfsdk:"occupied"`  // optional
}

type bedResource struct {
	provider provider
}

// compute head offset given a facing
func bedOffset(facing string) (dx, dz int, valid bool) {
	switch facing {
	case "north":
		return 0, -1, true // Z decreases to the north
	case "south":
		return 0, 1, true // Z increases to the south
	case "east":
		return 1, 0, true // X increases to the east
	case "west":
		return -1, 0, true // X decreases to the west
	default:
		return 0, 0, false
	}
}

func (r bedResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data bedResourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dx, dz, ok := bedOffset(data.Direction)
	if !ok {
		resp.Diagnostics.AddError("Validation Error", "direction must be one of north|south|east|west")
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	occupied := false
	if data.Occupied != nil {
		occupied = *data.Occupied
	}

	// Place FOOT at start position
	footMat := fmt.Sprintf(`%s[facing=%s,part=foot,occupied=%t]`, data.Material, data.Direction, occupied)
	if err := client.CreateBlock(ctx, footMat, data.Position.X, data.Position.Y, data.Position.Z); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place bed foot: %s", err))
		return
	}

	// Place HEAD one block in facing direction
	headX := data.Position.X + dx
	headZ := data.Position.Z + dz
	headMat := fmt.Sprintf(`%s[facing=%s,part=head,occupied=%t]`, data.Material, data.Direction, occupied)
	if err := client.CreateBlock(ctx, headMat, headX, data.Position.Y, headZ); err != nil {
		// Roll back foot on failure
		_ = client.DeleteBlock(ctx, data.Position.X, data.Position.Y, data.Position.Z)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place bed head: %s", err))
		return
	}

	data.Id = types.String{Value: fmt.Sprintf("bed-%d-%d-%d-%s", data.Position.X, data.Position.Y, data.Position.Z, data.Direction)}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r bedResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// No read API; keep state as-is
	var data bedResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r bedResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data bedResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dx, dz, ok := bedOffset(data.Direction)
	if !ok {
		resp.Diagnostics.AddError("Validation Error", "direction must be one of north|south|east|west")
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	occupied := false
	if data.Occupied != nil {
		occupied = *data.Occupied
	}

	// Re-place both parts
	footMat := fmt.Sprintf(`%s[facing=%s,part=foot,occupied=%t]`, data.Material, data.Direction, occupied)
	if err := client.CreateBlock(ctx, footMat, data.Position.X, data.Position.Y, data.Position.Z); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update bed foot: %s", err))
		return
	}

	headX := data.Position.X + dx
	headZ := data.Position.Z + dz
	headMat := fmt.Sprintf(`%s[facing=%s,part=head,occupied=%t]`, data.Material, data.Direction, occupied)
	if err := client.CreateBlock(ctx, headMat, headX, data.Position.Y, headZ); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update bed head: %s", err))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r bedResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data bedResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dx, dz, ok := bedOffset(data.Direction)
	if !ok {
		// Even if invalid, at least delete the foot
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	// Delete foot
	_ = client.DeleteBlock(ctx, data.Position.X, data.Position.Y, data.Position.Z)

	// Delete head (based on stored direction)
	if ok {
		headX := data.Position.X + dx
		headZ := data.Position.Z + dz
		_ = client.DeleteBlock(ctx, headX, data.Position.Y, headZ)
	}
}

func (r bedResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
