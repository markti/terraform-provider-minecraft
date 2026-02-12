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

var _ resource.Resource = (*stairsResource)(nil)
var _ resource.ResourceWithImportState = (*stairsResource)(nil)

type stairsResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Material types.String `tfsdk:"material"`
	Position struct {
		X types.Int32 `tfsdk:"x"`
		Y types.Int32 `tfsdk:"y"`
		Z types.Int32 `tfsdk:"z"`
	} `tfsdk:"position"`
	Facing      types.String `tfsdk:"facing"`
	Half        types.String `tfsdk:"half"`
	Shape       types.String `tfsdk:"shape"`
	Waterlogged types.Bool   `tfsdk:"waterlogged"`
}

type stairsResource struct {
	client *minecraft.Client
}

func NewStairsResource() resource.Resource {
	return &stairsResource{}
}

func (r *stairsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stairs"
}

func (r *stairsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Minecraft stairs block (e.g., minecraft:oak_stairs) with orientation and shape.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the stairs block.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"material": schema.StringAttribute{
				MarkdownDescription: "The stairs material (e.g., `minecraft:oak_stairs`, `minecraft:stone_brick_stairs`).",
				Required:            true,
			},
			"position": schema.SingleNestedAttribute{
				MarkdownDescription: "The position of the stairs block.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x": schema.Int32Attribute{
						MarkdownDescription: "X coordinate of the block",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"y": schema.Int32Attribute{
						MarkdownDescription: "Y coordinate of the block",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"z": schema.Int32Attribute{
						MarkdownDescription: "Z coordinate of the block",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"facing": schema.StringAttribute{
				MarkdownDescription: "Direction the stairs face: one of `north`, `south`, `east`, `west`.",
				Required:            true,
			},
			"half": schema.StringAttribute{
				MarkdownDescription: "Whether the stairs are on the `top` (upside-down) or `bottom` half.",
				Required:            true,
			},
			"shape": schema.StringAttribute{
				MarkdownDescription: "Stair shape: `straight`, `inner_left`, `inner_right`, `outer_left`, or `outer_right`.",
				Required:            true,
			},
			"waterlogged": schema.BoolAttribute{
				MarkdownDescription: "Whether the stairs are waterlogged. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *stairsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *stairsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data stairsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateStairs(
		ctx,
		data.Material.ValueString(),
		int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32()),
		data.Facing.ValueString(),
		data.Half.ValueString(),
		data.Shape.ValueString(),
		data.Waterlogged.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create stairs, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("stairs-%d-%d-%d", data.Position.X.ValueInt32(), data.Position.Y.ValueInt32(), data.Position.Z.ValueInt32()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *stairsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data stairsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *stairsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data stairsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateStairs(
		ctx,
		data.Material.ValueString(),
		int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32()),
		data.Facing.ValueString(),
		data.Half.ValueString(),
		data.Shape.ValueString(),
		data.Waterlogged.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update stairs, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *stairsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data stairsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBlock(ctx, int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete stairs, got error: %s", err))
		return
	}
}

func (r *stairsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

