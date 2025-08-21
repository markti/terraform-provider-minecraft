package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure framework interfaces
var _ tfsdk.ResourceType = teamMemberResourceType{}
var _ tfsdk.Resource = teamMemberResource{}
var _ tfsdk.ResourceWithImportState = teamMemberResource{}

// ----- Resource Type -----

type teamMemberResourceType struct{}

func (t teamMemberResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Adds a single player/selector/entity to a Minecraft team and removes it on destroy.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Composite ID: `team|kind|value` (e.g., `blue|player|Steve`).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"team": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "Target team name to join.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // changing team => recreate
				},
			},

			// Exactly ONE of the following must be set:
			"player": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Minecraft player username to add to the team.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"selector": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Target selector string (e.g. `@a[team=]`, `@e[type=minecraft:zombie,limit=1]`).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"entity_id": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Exact CustomName (text component string value) of the entity to add (e.g., a UUID you set when summoning).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
		},
	}, nil
}

func (t teamMemberResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return teamMemberResource{provider: p}, diags
}

// ----- Data Model -----

type teamMemberData struct {
	ID       types.String `tfsdk:"id"`
	Team     types.String `tfsdk:"team"`
	Player   types.String `tfsdk:"player"`
	Selector types.String `tfsdk:"selector"`
	EntityID types.String `tfsdk:"entity_id"`
}

type teamMemberResource struct {
	provider provider
}

// ----- CRUD -----

func (r teamMemberResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan teamMemberData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kind, val, err := validateAndPickTarget(plan, &resp.Diagnostics)
	if err != nil {
		return
	}

	client, err2 := r.provider.GetClient(ctx)
	if err2 != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err2))
		return
	}

	team := strings.TrimSpace(plan.Team.Value)

	switch kind {
	case "player":
		if err := client.JoinTeamPlayers(ctx, team, val); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add player %q to team %q: %s", val, team, err))
			return
		}
	case "selector":
		if err := client.JoinTeamTargets(ctx, team, val); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add selector %q to team %q: %s", val, team, err))
			return
		}
	case "entity":
		if err := client.JoinTeamEntityByName(ctx, team, val); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add entity %q to team %q: %s", val, team, err))
			return
		}
	default:
		resp.Diagnostics.AddError("Validation Error", "unknown membership kind")
		return
	}

	plan.ID = types.String{Value: fmt.Sprintf("%s|%s|%s", team, kind, val)}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r teamMemberResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// No reliable query for membership by player/selector/entity via RCON without heavy parsing.
	// Keep state as-is (best-effort). You can implement drift detection later by parsing `/team list <team>`.
	var state teamMemberData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r teamMemberResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// All fields RequireReplace; nothing to update in place.
	var plan teamMemberData
	_ = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.AddWarning("No-op Update", "All attributes are ForceNew; resource will be recreated if changed.")
	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r teamMemberResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state teamMemberData
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

	kind, val, _ := validateAndPickTarget(state, &resp.Diagnostics)
	// Even if validate returns an error, try best-effort cleanup based on ID.
	if kind == "" || val == "" {
		kind, val = parseIDFallback(state.ID.Value)
	}

	switch kind {
	case "player":
		if err := client.LeaveTeamPlayers(ctx, val); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove player %q from team: %s", val, err))
		}
	case "selector":
		if err := client.LeaveTeamTargets(ctx, val); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove selector %q from team: %s", val, err))
		}
	case "entity":
		if err := client.LeaveTeamEntityByName(ctx, val); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove entity %q from team: %s", val, err))
		}
	default:
		// Nothing we can do
	}
}

func (r teamMemberResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Expect ID in the form: team|kind|value
	parts := strings.SplitN(req.ID, "|", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Import Error", "Expected ID in format `team|kind|value` (e.g., `blue|player|Steve`).")
		return
	}
	team, kind, value := parts[0], parts[1], parts[2]

	var st teamMemberData
	st.ID = types.String{Value: req.ID}
	st.Team = types.String{Value: team}

	switch kind {
	case "player":
		st.Player = types.String{Value: value}
	case "selector":
		st.Selector = types.String{Value: value}
	case "entity":
		st.EntityID = types.String{Value: value}
	default:
		resp.Diagnostics.AddError("Import Error", "kind must be one of `player`, `selector`, or `entity`.")
		return
	}

	diags := resp.State.Set(ctx, &st)
	resp.Diagnostics.Append(diags...)
}

// ----- Helpers -----

func validateAndPickTarget(d teamMemberData, diags *diag.Diagnostics) (kind string, value string, err error) {
	team := strings.TrimSpace(d.Team.Value)
	if team == "" {
		diags.AddError("Validation Error", "`team` is required.")
		return "", "", fmt.Errorf("team required")
	}

	cnt := 0
	if !d.Player.Null && strings.TrimSpace(d.Player.Value) != "" {
		cnt++
		kind = "player"
		value = strings.TrimSpace(d.Player.Value)
	}
	if !d.Selector.Null && strings.TrimSpace(d.Selector.Value) != "" {
		cnt++
		kind = "selector"
		value = strings.TrimSpace(d.Selector.Value)
	}
	if !d.EntityID.Null && strings.TrimSpace(d.EntityID.Value) != "" {
		cnt++
		kind = "entity"
		value = strings.TrimSpace(d.EntityID.Value)
	}

	if cnt == 0 {
		diags.AddError("Validation Error", "Exactly one of `player`, `selector`, or `entity_id` must be set.")
		return "", "", fmt.Errorf("no target")
	}
	if cnt > 1 {
		diags.AddError("Validation Error", "Only one of `player`, `selector`, or `entity_id` may be set.")
		return "", "", fmt.Errorf("multiple targets")
	}
	return kind, value, nil
}

func parseIDFallback(id string) (kind, value string) {
	parts := strings.SplitN(id, "|", 3)
	if len(parts) == 3 {
		return parts[1], parts[2]
	}
	return "", ""
}
