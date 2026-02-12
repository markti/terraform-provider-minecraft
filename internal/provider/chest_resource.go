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

var _ resource.Resource = (*chestResource)(nil)
var _ resource.ResourceWithImportState = (*chestResource)(nil)

type chestResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Size        types.String `tfsdk:"size"`
	Trapped     types.Bool   `tfsdk:"trapped"`
	Waterlogged types.Bool   `tfsdk:"waterlogged"`
	Position    struct {
		X types.Int32 `tfsdk:"x"`
		Y types.Int32 `tfsdk:"y"`
		Z types.Int32 `tfsdk:"z"`
	} `tfsdk:"position"`
}

type chestResource struct {
	client *minecraft.Client
}

func NewChestResource() resource.Resource {
	return &chestResource{}
}

func (r *chestResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chest"
}

func (r *chestResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Minecraft chest. Can be a single chest or a double chest (two blocks side by side).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the chest resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"position": schema.SingleNestedAttribute{
				MarkdownDescription: "The position of the first chest block.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x": schema.Int32Attribute{
						Required:            true,
						MarkdownDescription: "X coordinate",
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"y": schema.Int32Attribute{
						Required:            true,
						MarkdownDescription: "Y coordinate",
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"z": schema.Int32Attribute{
						Required:            true,
						MarkdownDescription: "Z coordinate",
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"size": schema.StringAttribute{
				MarkdownDescription: "The chest size: `single` or `double`.",
				Required:            true,
			},
			"trapped": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a trapped chest. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"waterlogged": schema.BoolAttribute{
				MarkdownDescription: "Whether the chest is waterlogged. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *chestResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *chestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data chestResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	material := "minecraft:chest"
	if data.Trapped.ValueBool() {
		material = "minecraft:trapped_chest"
	}

	switch data.Size.ValueString() {
	case "single":
		block := fmt.Sprintf(`%s[type=single,waterlogged=%t]`, material, data.Waterlogged.ValueBool())
		if err := r.client.CreateBlock(ctx, block, int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32())); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place single chest: %s", err))
			return
		}
	case "double":
		blockLeft := fmt.Sprintf(`%s[type=left,waterlogged=%t]`, material, data.Waterlogged.ValueBool())
		blockRight := fmt.Sprintf(`%s[type=right,waterlogged=%t]`, material, data.Waterlogged.ValueBool())
		x := int(data.Position.X.ValueInt32())
		y := int(data.Position.Y.ValueInt32())
		z := int(data.Position.Z.ValueInt32())
		if err := r.client.CreateBlock(ctx, blockLeft, x, y, z); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place left half of double chest: %s", err))
			return
		}
		if err := r.client.CreateBlock(ctx, blockRight, x+1, y, z); err != nil {
			_ = r.client.DeleteBlock(ctx, x, y, z)
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place right half of double chest: %s", err))
			return
		}
	default:
		resp.Diagnostics.AddError("Validation Error", "size must be 'single' or 'double'")
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("chest-%d-%d-%d", data.Position.X.ValueInt32(), data.Position.Y.ValueInt32(), data.Position.Z.ValueInt32()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *chestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data chestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *chestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data chestResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	material := "minecraft:chest"
	if data.Trapped.ValueBool() {
		material = "minecraft:trapped_chest"
	}

	switch data.Size.ValueString() {
	case "single":
		block := fmt.Sprintf(`%s[type=single,waterlogged=%t]`, material, data.Waterlogged.ValueBool())
		if err := r.client.CreateBlock(ctx, block, int(data.Position.X.ValueInt32()), int(data.Position.Y.ValueInt32()), int(data.Position.Z.ValueInt32())); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update single chest: %s", err))
			return
		}
	case "double":
		blockLeft := fmt.Sprintf(`%s[type=left,waterlogged=%t]`, material, data.Waterlogged.ValueBool())
		blockRight := fmt.Sprintf(`%s[type=right,waterlogged=%t]`, material, data.Waterlogged.ValueBool())
		x := int(data.Position.X.ValueInt32())
		y := int(data.Position.Y.ValueInt32())
		z := int(data.Position.Z.ValueInt32())
		if err := r.client.CreateBlock(ctx, blockLeft, x, y, z); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update left half of double chest: %s", err))
			return
		}
		if err := r.client.CreateBlock(ctx, blockRight, x+1, y, z); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update right half of double chest: %s", err))
			return
		}
	default:
		resp.Diagnostics.AddError("Validation Error", "size must be 'single' or 'double'")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *chestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data chestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	x := int(data.Position.X.ValueInt32())
	y := int(data.Position.Y.ValueInt32())
	z := int(data.Position.Z.ValueInt32())

	_ = r.client.DeleteBlock(ctx, x, y, z)
	if data.Size.ValueString() == "double" {
		_ = r.client.DeleteBlock(ctx, x+1, y, z)
	}
}

func (r *chestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

