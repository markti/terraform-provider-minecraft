package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure types satisfy framework interfaces
var _ tfsdk.ResourceType = teamResourceType{}
var _ tfsdk.Resource = teamResource{}
var _ tfsdk.ResourceWithImportState = teamResource{}

// -------- Resource Type --------

type teamResourceType struct{}

func (t teamResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A Minecraft scoreboard team managed via RCON.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Resource ID (same as `name`).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"name": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "Team name (identifier).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // renaming team => ForceNew
				},
			},
			"display_name": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Display name shown in UI (defaults to `name`).",
			},
			"color": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Team color (e.g. `red`, `blue`, `gold`, `dark_purple`, etc.).",
			},
			"friendly_fire": {
				Type:                types.BoolType,
				Optional:            true,
				MarkdownDescription: "Whether teammates can damage each other.",
			},
			"see_friendly_invisibles": {
				Type:                types.BoolType,
				Optional:            true,
				MarkdownDescription: "If true, teammates can see each other when invisible.",
			},
			"nametag_visibility": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "One of `always`, `never`, `hideForOtherTeams`, `hideForOwnTeam`.",
			},
			"collision_rule": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "One of `always`, `never`, `pushOtherTeams`, `pushOwnTeam`.",
			},
		},
	}, nil
}

func (t teamResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return teamResource{provider: p}, diags
}

// -------- Data & Resource --------

type teamResourceData struct {
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
	provider provider
}

// -------- CRUD --------

func (r teamResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan teamResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	name := strings.TrimSpace(plan.Name.Value)
	display := name
	if !plan.DisplayName.Null && plan.DisplayName.Value != "" {
		display = plan.DisplayName.Value
	}

	// Create team
	if err := client.CreateTeam(ctx, name, display); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team: %s", err))
		return
	}

	// Apply options present in plan
	if err := applyTeamOptions(ctx, client, name, plan, &resp.Diagnostics); err != nil {
		return
	}

	plan.ID = types.String{Value: name}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r teamResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Minimal read; keep state as-is. (Add drift detection later by parsing `/team list`.)
	var state teamResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r teamResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state teamResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	name := strings.TrimSpace(plan.Name.Value)

	// display_name change
	if !equalString(plan.DisplayName, state.DisplayName) {
		display := name
		if !plan.DisplayName.Null && plan.DisplayName.Value != "" {
			display = plan.DisplayName.Value
		}
		if err := client.SetTeamDisplayName(ctx, name, display); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set displayName: %s", err))
			return
		}
	}

	// Apply (or re-apply) the rest of the options
	if err := applyTeamOptions(ctx, client, name, plan, &resp.Diagnostics); err != nil {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r teamResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state teamResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	if err := client.DeleteTeam(ctx, state.Name.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team: %s", err))
		return
	}
}

func (r teamResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Import by team name into `id`; user config supplies `name`.
	// (Or you can set both name and id here if you prefer strict import.)
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

// -------- Helpers --------

func equalString(a, b types.String) bool {
	if a.Null && b.Null {
		return true
	}
	return a.Value == b.Value
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

func applyTeamOptions(ctx context.Context, c teamOptionClient, name string, d teamResourceData, diags *diag.Diagnostics) error {
	// color
	if !d.Color.Null && d.Color.Value != "" {
		if err := c.SetTeamColor(ctx, name, strings.ToLower(d.Color.Value)); err != nil {
			diags.AddError("Client Error", fmt.Sprintf("Unable to set color: %s", err))
			return err
		}
	}
	// friendlyFire
	if !d.FriendlyFire.Null {
		if err := c.SetTeamFriendlyFire(ctx, name, d.FriendlyFire.Value); err != nil {
			diags.AddError("Client Error", fmt.Sprintf("Unable to set friendlyFire: %s", err))
			return err
		}
	}
	// seeFriendlyInvisibles
	if !d.SeeFriendlyInvisibles.Null {
		if err := c.SetTeamSeeFriendlyInvisibles(ctx, name, d.SeeFriendlyInvisibles.Value); err != nil {
			diags.AddError("Client Error", fmt.Sprintf("Unable to set seeFriendlyInvisibles: %s", err))
			return err
		}
	}
	// nametagVisibility
	if !d.NametagVisibility.Null && d.NametagVisibility.Value != "" {
		if err := c.SetTeamNametagVisibility(ctx, name, d.NametagVisibility.Value); err != nil {
			diags.AddError("Client Error", fmt.Sprintf("Unable to set nametagVisibility: %s", err))
			return err
		}
	}
	// collisionRule
	if !d.CollisionRule.Null && d.CollisionRule.Value != "" {
		if err := c.SetTeamCollisionRule(ctx, name, d.CollisionRule.Value); err != nil {
			diags.AddError("Client Error", fmt.Sprintf("Unable to set collisionRule: %s", err))
			return err
		}
	}
	return nil
}
