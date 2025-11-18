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
var _ tfsdk.ResourceType = zombieResourceType{}
var _ tfsdk.Resource = zombieResource{}
var _ tfsdk.ResourceWithImportState = zombieResource{}

// ---------- Resource Type ----------

type zombieResourceType struct{}

func (t zombieResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Summon and manage a Minecraft zombie with baby/door-breaking/loot/persistence options.",
		Attributes: map[string]tfsdk.Attribute{
			"position": {
				MarkdownDescription: "Where to summon the zombie.",
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
			"is_baby": {
				MarkdownDescription: "Whether the zombie is a baby. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Type:                types.BoolType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"can_break_doors": {
				MarkdownDescription: "Whether the zombie can break doors. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Type:                types.BoolType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"can_pick_up_loot": {
				MarkdownDescription: "Whether the zombie can pick up loot. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Type:                types.BoolType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"persistence_required": {
				MarkdownDescription: "Whether the zombie is prevented from naturally despawning. Defaults to `false` if not set.",
				Optional:            true,
				Computed:            true,
				Type:                types.BoolType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"health": {
				MarkdownDescription: "Zombie health (float). Defaults to `20.0` if not set.",
				Optional:            true,
				Computed:            true,
				Type:                types.Float64Type,
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

func (t zombieResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return zombieResource{provider: p}, diags
}

// ---------- Resource Data ----------

type zombieResourceData struct {
	Id       types.String `tfsdk:"id"`
	Position struct {
		X int64 `tfsdk:"x"`
		Y int64 `tfsdk:"y"`
		Z int64 `tfsdk:"z"`
	} `tfsdk:"position"`

	IsBaby             types.Bool   `tfsdk:"is_baby"`
	CanBreakDoors      types.Bool   `tfsdk:"can_break_doors"`
	CanPickUpLoot      types.Bool   `tfsdk:"can_pick_up_loot"`
	PersistenceRequired types.Bool  `tfsdk:"persistence_required"`
	Health             types.Float64 `tfsdk:"health"`
}

// ---------- Resource Impl ----------

type zombieResource struct {
	provider provider
}

func (r zombieResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data zombieResourceData
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

	// Default bools to false when null/unknown
	if data.IsBaby.Null || data.IsBaby.Unknown {
		data.IsBaby = types.Bool{Value: false}
	}
	if data.CanBreakDoors.Null || data.CanBreakDoors.Unknown {
		data.CanBreakDoors = types.Bool{Value: false}
	}
	if data.CanPickUpLoot.Null || data.CanPickUpLoot.Unknown {
		data.CanPickUpLoot = types.Bool{Value: false}
	}
	if data.PersistenceRequired.Null || data.PersistenceRequired.Unknown {
		data.PersistenceRequired = types.Bool{Value: false}
	}

	// Default health to full (20.0) when null/unknown
	if data.Health.Null || data.Health.Unknown {
		data.Health = types.Float64{Value: 20.0}
	}

	id := uuid.NewString()
	pos := fmt.Sprintf("%d %d %d", data.Position.X, data.Position.Y, data.Position.Z)

	// Use the specialized client method to include zombie-specific NBT
	if err := client.CreateZombie(
		ctx,
		pos,
		id,
		data.IsBaby.Value,
		data.CanBreakDoors.Value,
		data.CanPickUpLoot.Value,
		data.PersistenceRequired.Value,
		float32(data.Health.Value),
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to summon zombie: %s", err))
		return
	}

	data.Id = types.String{Value: id}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r zombieResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data zombieResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// No live read yet; just persist current state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r zombieResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data zombieResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// All attributes are ForceNew; no in-place update
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r zombieResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data zombieResourceData
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
	if err := client.DeleteEntity(ctx, "minecraft:zombie", pos, data.Id.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete zombie: %s", err))
		return
	}
}

func (r zombieResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Import by UUID (id). Config must specify matching position and attributes.
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
