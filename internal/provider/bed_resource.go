package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ resource.Resource = (*bedResource)(nil)
var _ resource.ResourceWithImportState = (*bedResource)(nil)

type bedResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Material types.String `tfsdk:"material"`
	Position struct {
		X types.Int32 `tfsdk:"x"`
		Y types.Int32 `tfsdk:"y"`
		Z types.Int32 `tfsdk:"z"`
	} `tfsdk:"position"`
	Direction types.String `tfsdk:"direction"`
	Occupied  types.Bool   `tfsdk:"occupied"`
}

type bedResource struct {
	client *minecraft.Client
}

func NewBedResource() resource.Resource {
	return &bedResource{}
}

func (r *bedResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bed"
}

func (r *bedResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Minecraft bed (two-block structure). The start position is the FOOT. Direction places the HEAD one block in that direction.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the bed resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"material": schema.StringAttribute{
				MarkdownDescription: "The bed material, e.g. `minecraft:red_bed`, `minecraft:blue_bed`.",
				Required:            true,
			},
			"position": schema.SingleNestedAttribute{
				MarkdownDescription: "The FOOT position of the bed.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x": schema.Int32Attribute{
						MarkdownDescription: "X coordinate (foot)",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"y": schema.Int32Attribute{
						MarkdownDescription: "Y coordinate",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"z": schema.Int32Attribute{
						MarkdownDescription: "Z coordinate (foot)",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"direction": schema.StringAttribute{
				MarkdownDescription: "Direction the bed faces: one of `north`, `south`, `east`, `west`. The HEAD goes one block in this direction from the FOOT.",
				Required:            true,
			},
			"occupied": schema.BoolAttribute{
				MarkdownDescription: "Whether the bed is considered occupied (rarely needed). Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *bedResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*minecraft.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *minecraft.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *bedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bedResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dx, dz, ok := bedOffset(data.Direction.ValueString())
	if !ok {
		resp.Diagnostics.AddError("Validation Error", "direction must be one of north|south|east|west")
		return
	}

	occupied := data.Occupied.ValueBool()

	// Place FOOT at start position
	footMat := fmt.Sprintf(`%s[facing=%s,part=foot,occupied=%t]`, data.Material.ValueString(), data.Direction.ValueString(), occupied)
	if err := r.client.CreateBlock(ctx, footMat, int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32())); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place bed foot: %s", err))
		return
	}

	// Place HEAD one block in facing direction
	headX := int(data.Position.X.ValueInt32()) + dx
	headZ := int(data.Position.Z.ValueInt32()) + dz
	headMat := fmt.Sprintf(`%s[facing=%s,part=head,occupied=%t]`, data.Material.ValueString(), data.Direction.ValueString(), occupied)
	if err := r.client.CreateBlock(ctx, headMat, headX, int(data.Position.Y.ValueInt32()), headZ); err != nil {
		// Roll back foot on failure
		_ = r.client.DeleteBlock(ctx, int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place bed head: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("bed-%d-%d-%d-%s", data.Position.X.ValueInt32(), data.Position.Y.ValueInt32(), data.Position.Z.ValueInt32(), data.Direction.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bedResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data bedResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dx, dz, ok := bedOffset(data.Direction.ValueString())
	if !ok {
		resp.Diagnostics.AddError("Validation Error", "direction must be one of north|south|east|west")
		return
	}

	occupied := data.Occupied.ValueBool()

	// Re-place both parts
	footMat := fmt.Sprintf(`%s[facing=%s,part=foot,occupied=%t]`, data.Material.ValueString(), data.Direction.ValueString(), occupied)
	if err := r.client.CreateBlock(ctx, footMat, int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32())); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update bed foot: %s", err))
		return
	}

	headX := int(data.Position.X.ValueInt32()) + dx
	headZ := int(data.Position.Z.ValueInt32()) + dz
	headMat := fmt.Sprintf(`%s[facing=%s,part=head,occupied=%t]`, data.Material.ValueString(), data.Direction.ValueString(), occupied)
	if err := r.client.CreateBlock(ctx, headMat, headX, int(data.Position.Y.ValueInt32()), headZ); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update bed head: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bedResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dx, dz, ok := bedOffset(data.Direction.ValueString())

	// Delete foot
	_ = r.client.DeleteBlock(ctx, int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32()))

	// Delete head (based on stored direction)
	if ok {
		headX := int(data.Position.X.ValueInt32()) + dx
		headZ := int(data.Position.Z.ValueInt32()) + dz
		_ = r.client.DeleteBlock(ctx, headX, int(data.Position.Y.ValueInt32()), headZ)
	}
}

func (r *bedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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

