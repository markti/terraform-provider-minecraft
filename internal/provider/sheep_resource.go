package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ resource.Resource = (*sheepResource)(nil)
var _ resource.ResourceWithImportState = (*sheepResource)(nil)

type sheepResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Color    types.String `tfsdk:"color"`
	Sheared  types.Bool   `tfsdk:"sheared"`
	Position struct {
		X types.Int32 `tfsdk:"x"`
		Y types.Int32 `tfsdk:"y"`
		Z types.Int32 `tfsdk:"z"`
	} `tfsdk:"position"`
}

type sheepResource struct {
	client *minecraft.Client
}

func NewSheepResource() resource.Resource {
	return &sheepResource{}
}

func (r *sheepResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sheep"
}

func (r *sheepResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Summon and manage a Minecraft sheep with color and sheared state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Stable UUID used as the entity's CustomName/tag.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"position": schema.SingleNestedAttribute{
				MarkdownDescription: "Where to summon the sheep.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x": schema.Int32Attribute{
						MarkdownDescription: "X coordinate",
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
						MarkdownDescription: "Z coordinate",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "Sheep wool color (string).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sheared": schema.BoolAttribute{
				MarkdownDescription: "Whether the sheep starts sheared. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *sheepResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sheepResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data sheepResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.NewString()
	pos := fmt.Sprintf("%d %d %d", data.Position.X.ValueInt32(), data.Position.Y.ValueInt32(), data.Position.Z.ValueInt32())
	if err := r.client.CreateSheep(ctx, pos, id, strings.ToLower(data.Color.ValueString()), data.Sheared.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to summon sheep: %s", err))
		return
	}

	data.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sheepResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data sheepResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sheepResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data sheepResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sheepResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data sheepResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pos := fmt.Sprintf("%d %d %d", data.Position.X.ValueInt32(), data.Position.Y.ValueInt32(), data.Position.Z.ValueInt32())
	if err := r.client.DeleteEntity(ctx, "minecraft:sheep", pos, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete sheep: %s", err))
		return
	}
}

func (r *sheepResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

