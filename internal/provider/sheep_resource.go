package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = sheepResourceType{}
var _ tfsdk.Resource = sheepResource{}
var _ tfsdk.ResourceWithImportState = sheepResource{}

// ---------- Resource Type ----------

type sheepResourceType struct{}

func (t sheepResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Summon and manage a Minecraft sheep with color and sheared state.",
		Attributes: map[string]tfsdk.Attribute{
			"position": {
				MarkdownDescription: "Where to summon the sheep.",
				Required:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"x": {
						MarkdownDescription: "X coordinate",
						Type:                types.Int64Type,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"y": {
						MarkdownDescription: "Y coordinate",
						Type:                types.Int64Type,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"z": {
						MarkdownDescription: "Z coordinate",
						Type:                types.Int64Type,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
				}),
			},
			"color": {
				MarkdownDescription: "Sheep wool color (string). One of: `white, orange, magenta, light_blue, yellow, lime, pink, gray, light_gray, cyan, purple, blue, brown, green, red, black`.",
				Required: true,
				Type:     types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"sheared": {
				MarkdownDescription: "Whether the sheep starts sheared. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true, // lets us keep state = false and avoid unknowns
				Type:                types.BoolType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Stable UUID used as the entity's CustomName/tag.",
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (t sheepResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return sheepResource{provider: p}, diags
}

// ---------- Resource Data ----------

type sheepResourceData struct {
	Id       types.String `tfsdk:"id"`
	Position struct {
		X int64 `tfsdk:"x"`
		Y int64 `tfsdk:"y"`
		Z int64 `tfsdk:"z"`
	} `tfsdk:"position"`
	Color   string     `tfsdk:"color"`
	Sheared types.Bool `tfsdk:"sheared"`
}

// ---------- Resource Impl ----------

type sheepResource struct {
	provider provider
}

func (r sheepResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data sheepResourceData
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

	// Default sheared = false when null/unknown
	if data.Sheared.Null || data.Sheared.Unknown {
		data.Sheared = types.Bool{Value: false}
	}

	id := uuid.NewString()
	pos := fmt.Sprintf("%d %d %d", data.Position.X, data.Position.Y, data.Position.Z)

	// Use the specialized client method to include sheep-specific NBT
	if err := client.CreateSheep(ctx, pos, id, strings.ToLower(data.Color), data.Sheared.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to summon sheep: %s", err))
		return
	}

	data.Id = types.String{Value: id}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r sheepResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data sheepResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data) // no live read yet
	resp.Diagnostics.Append(diags...)
}

func (r sheepResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data sheepResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data) // all fields ForceNew; nothing in-place
	resp.Diagnostics.Append(diags...)
}

func (r sheepResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data sheepResourceData
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
	if err := client.DeleteEntity(ctx, "minecraft:sheep", pos, data.Id.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete sheep: %s", err))
		return
	}
}

func (r sheepResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Import by UUID (id). Config must specify matching position/color/sheared.
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
