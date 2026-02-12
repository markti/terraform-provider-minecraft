package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ resource.Resource = (*teamResource)(nil)
var _ resource.ResourceWithImportState = (*teamResource)(nil)

type teamResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	DisplayName           types.String `tfsdk:"display_name"`
	Color                 types.String `tfsdk:"color"`
	FriendlyFire          types.Bool   `tfsdk:"friendly_fire"`
	SeeFriendlyInvisibles types.Bool   `tfsdk:"see_friendly_invisibles"`
	NametagVisibility     types.String `tfsdk:"nametag_visibility"`
	CollisionRule         types.String `tfsdk:"collision_rule"`
}

type teamResource struct {
	client *minecraft.Client
}

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

func (r *teamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Minecraft scoreboard team managed via RCON.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource ID (same as `name`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Team name (identifier).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name shown in UI (defaults to `name`).",
				Optional:            true,
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "Team color (e.g. `red`, `blue`, `gold`, `dark_purple`, etc.).",
				Optional:            true,
			},
			"friendly_fire": schema.BoolAttribute{
				MarkdownDescription: "Whether teammates can damage each other.",
				Optional:            true,
			},
			"see_friendly_invisibles": schema.BoolAttribute{
				MarkdownDescription: "If true, teammates can see each other when invisible.",
				Optional:            true,
			},
			"nametag_visibility": schema.StringAttribute{
				MarkdownDescription: "One of `always`, `never`, `hideForOtherTeams`, `hideForOwnTeam`.",
				Optional:            true,
			},
			"collision_rule": schema.StringAttribute{
				MarkdownDescription: "One of `always`, `never`, `pushOtherTeams`, `pushOwnTeam`.",
				Optional:            true,
			},
		},
	}
}

func (r *teamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(data.Name.ValueString())
	if name == "" {
		resp.Diagnostics.AddError("Validation Error", "Attribute `name` cannot be empty or whitespace.")
		return
	}

	display := name
	if !data.DisplayName.IsNull() && !data.DisplayName.IsUnknown() {
		if trimmed := strings.TrimSpace(data.DisplayName.ValueString()); trimmed != "" {
			display = trimmed
		}
	}

	if err := r.client.CreateTeam(ctx, name, display); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team: %s", err))
		return
	}

	if err := applyTeamOptions(ctx, r.client, name, data); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	data.ID = types.StringValue(name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var plan, state teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(plan.Name.ValueString())
	if name == "" {
		resp.Diagnostics.AddError("Validation Error", "Attribute `name` cannot be empty or whitespace.")
		return
	}

	if !stringValuesEqual(plan.DisplayName, state.DisplayName) {
		display := name
		if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
			if trimmed := strings.TrimSpace(plan.DisplayName.ValueString()); trimmed != "" {
				display = trimmed
			}
		}
		if err := r.client.SetTeamDisplayName(ctx, name, display); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set displayName: %s", err))
			return
		}
	}

	if err := applyTeamOptions(ctx, r.client, name, plan); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	if plan.ID.IsNull() || plan.ID.IsUnknown() {
		plan.ID = types.StringValue(name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(state.Name.ValueString())
	if name == "" {
		name = strings.TrimSpace(state.ID.ValueString())
	}
	if name == "" {
		resp.Diagnostics.AddError("Validation Error", "Missing team name for delete.")
		return
	}

	if err := r.client.DeleteTeam(ctx, name); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team: %s", err))
		return
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := strings.TrimSpace(req.ID)
	if name == "" {
		resp.Diagnostics.AddError("Import Error", "Expected non-empty team name as import ID.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

func stringValuesEqual(a, b types.String) bool {
	if a.IsNull() && b.IsNull() {
		return true
	}
	if a.IsUnknown() || b.IsUnknown() {
		return false
	}
	return a.ValueString() == b.ValueString()
}

type teamOptionClient interface {
	SetTeamDisplayName(ctx context.Context, name, display string) error
	SetTeamColor(ctx context.Context, name, color string) error
	SetTeamFriendlyFire(ctx context.Context, name string, enabled bool) error
	SetTeamSeeFriendlyInvisibles(ctx context.Context, name string, enabled bool) error
	SetTeamNametagVisibility(ctx context.Context, name, mode string) error
	SetTeamCollisionRule(ctx context.Context, name, rule string) error
	CreateTeam(ctx context.Context, name, display string) error
	DeleteTeam(ctx context.Context, name string) error
}

func applyTeamOptions(ctx context.Context, c teamOptionClient, name string, d teamResourceModel) error {
	if !d.Color.IsNull() && !d.Color.IsUnknown() && strings.TrimSpace(d.Color.ValueString()) != "" {
		if err := c.SetTeamColor(ctx, name, strings.ToLower(d.Color.ValueString())); err != nil {
			return fmt.Errorf("unable to set color: %w", err)
		}
	}
	if !d.FriendlyFire.IsNull() && !d.FriendlyFire.IsUnknown() {
		if err := c.SetTeamFriendlyFire(ctx, name, d.FriendlyFire.ValueBool()); err != nil {
			return fmt.Errorf("unable to set friendlyFire: %w", err)
		}
	}
	if !d.SeeFriendlyInvisibles.IsNull() && !d.SeeFriendlyInvisibles.IsUnknown() {
		if err := c.SetTeamSeeFriendlyInvisibles(ctx, name, d.SeeFriendlyInvisibles.ValueBool()); err != nil {
			return fmt.Errorf("unable to set seeFriendlyInvisibles: %w", err)
		}
	}
	if !d.NametagVisibility.IsNull() && !d.NametagVisibility.IsUnknown() && strings.TrimSpace(d.NametagVisibility.ValueString()) != "" {
		if err := c.SetTeamNametagVisibility(ctx, name, d.NametagVisibility.ValueString()); err != nil {
			return fmt.Errorf("unable to set nametagVisibility: %w", err)
		}
	}
	if !d.CollisionRule.IsNull() && !d.CollisionRule.IsUnknown() && strings.TrimSpace(d.CollisionRule.ValueString()) != "" {
		if err := c.SetTeamCollisionRule(ctx, name, d.CollisionRule.ValueString()); err != nil {
			return fmt.Errorf("unable to set collisionRule: %w", err)
		}
	}
	return nil
}

