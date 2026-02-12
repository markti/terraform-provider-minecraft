package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ resource.Resource = (*zombieResource)(nil)
var _ resource.ResourceWithImportState = (*zombieResource)(nil)

type zombieResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Position struct {
		X types.Int64 `tfsdk:"x"`
		Y types.Int64 `tfsdk:"y"`
		Z types.Int64 `tfsdk:"z"`
	} `tfsdk:"position"`
	IsBaby              types.Bool    `tfsdk:"is_baby"`
	CanBreakDoors       types.Bool    `tfsdk:"can_break_doors"`
	CanPickUpLoot       types.Bool    `tfsdk:"can_pick_up_loot"`
	PersistenceRequired types.Bool    `tfsdk:"persistence_required"`
	Health              types.Float64 `tfsdk:"health"`
}

type zombieResource struct {
	client *minecraft.Client
}

func NewZombieResource() resource.Resource {
	return &zombieResource{}
}

func (r *zombieResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zombie"
}

func (r *zombieResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Summon and manage a Minecraft zombie with baby/door-breaking/loot/persistence options.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Stable UUID used as the entity's CustomName/tag.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"position": schema.SingleNestedAttribute{
				MarkdownDescription: "Where to summon the zombie.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x": schema.Int64Attribute{
						MarkdownDescription: "X coordinate",
						Required:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
					"y": schema.Int64Attribute{
						MarkdownDescription: "Y coordinate",
						Required:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
					"z": schema.Int64Attribute{
						MarkdownDescription: "Z coordinate",
						Required:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"is_baby": schema.BoolAttribute{
				MarkdownDescription: "Whether the zombie is a baby. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"can_break_doors": schema.BoolAttribute{
				MarkdownDescription: "Whether the zombie can break doors. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"can_pick_up_loot": schema.BoolAttribute{
				MarkdownDescription: "Whether the zombie can pick up loot. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"persistence_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the zombie is prevented from naturally despawning. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"health": schema.Float64Attribute{
				MarkdownDescription: "Zombie health (float). Defaults to `20.0` if not set.",
				Optional:            true,
				Computed:            true,
				Default:             float64default.StaticFloat64(20.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *zombieResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *zombieResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data zombieResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := uuid.NewString()
	pos := fmt.Sprintf("%d %d %d", data.Position.X.ValueInt64(), data.Position.Y.ValueInt64(), data.Position.Z.ValueInt64())

	if err := r.client.CreateZombie(
		ctx,
		pos,
		id,
		data.IsBaby.ValueBool(),
		data.CanBreakDoors.ValueBool(),
		data.CanPickUpLoot.ValueBool(),
		data.PersistenceRequired.ValueBool(),
		float32(data.Health.ValueFloat64()),
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to summon zombie: %s", err))
		return
	}

	data.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *zombieResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data zombieResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *zombieResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data zombieResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *zombieResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data zombieResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pos := fmt.Sprintf("%d %d %d", data.Position.X.ValueInt64(), data.Position.Y.ValueInt64(), data.Position.Z.ValueInt64())
	if err := r.client.DeleteEntity(ctx, "minecraft:zombie", pos, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete zombie: %s", err))
		return
	}
}

func (r *zombieResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

