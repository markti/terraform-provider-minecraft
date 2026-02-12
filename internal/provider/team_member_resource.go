package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ resource.Resource = (*teamMemberResource)(nil)
var _ resource.ResourceWithImportState = (*teamMemberResource)(nil)

type teamMemberResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Team     types.String `tfsdk:"team"`
	Player   types.String `tfsdk:"player"`
	Selector types.String `tfsdk:"selector"`
	EntityID types.String `tfsdk:"entity_id"`
}

type teamMemberResource struct {
	client *minecraft.Client
}

func NewTeamMemberResource() resource.Resource {
	return &teamMemberResource{}
}

func (r *teamMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_member"
}

func (r *teamMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Adds a single player/selector/entity to a Minecraft team and removes it on destroy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite ID: `team|kind|value` (e.g., `blue|player|Steve`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team": schema.StringAttribute{
				MarkdownDescription: "Target team name to join.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"player": schema.StringAttribute{
				MarkdownDescription: "Minecraft player username to add to the team.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"selector": schema.StringAttribute{
				MarkdownDescription: "Target selector string (e.g. `@a[team=]`, `@e[type=minecraft:zombie,limit=1]`).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_id": schema.StringAttribute{
				MarkdownDescription: "Exact CustomName (text component string value) of the entity to add (e.g., a UUID you set when summoning).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *teamMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data teamMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, kind, value, err := validateTeamMemberCreate(data)
	if err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	switch kind {
	case "player":
		if err := r.client.JoinTeamPlayers(ctx, team, value); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add player %q to team %q: %s", value, team, err))
			return
		}
	case "selector":
		if err := r.client.JoinTeamTargets(ctx, team, value); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add selector %q to team %q: %s", value, team, err))
			return
		}
	case "entity":
		if err := r.client.JoinTeamEntityByName(ctx, team, value); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add entity %q to team %q: %s", value, team, err))
			return
		}
	default:
		resp.Diagnostics.AddError("Validation Error", "Unknown membership kind.")
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s|%s|%s", team, kind, value))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data teamMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	kind, value := pickTeamMemberTarget(data)
	if kind == "" || value == "" {
		kind, value = parseTeamMemberID(data.ID.ValueString())
	}
	if value == "" {
		return
	}

	switch kind {
	case "player":
		if err := r.client.LeaveTeamPlayers(ctx, value); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove player %q from team: %s", value, err))
		}
	case "selector":
		if err := r.client.LeaveTeamTargets(ctx, value); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove selector %q from team: %s", value, err))
		}
	case "entity":
		if err := r.client.LeaveTeamEntityByName(ctx, value); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove entity %q from team: %s", value, err))
		}
	default:
	}
}

func (r *teamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "|", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Import Error", "Expected ID in format `team|kind|value` (e.g., `blue|player|Steve`).")
		return
	}
	team, kind, value := parts[0], parts[1], parts[2]

	var st teamMemberResourceModel
	st.ID = types.StringValue(req.ID)
	st.Team = types.StringValue(team)

	switch kind {
	case "player":
		st.Player = types.StringValue(value)
	case "selector":
		st.Selector = types.StringValue(value)
	case "entity":
		st.EntityID = types.StringValue(value)
	default:
		resp.Diagnostics.AddError("Import Error", "kind must be one of `player`, `selector`, or `entity`.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}

func validateTeamMemberCreate(d teamMemberResourceModel) (team, kind, value string, err error) {
	team = strings.TrimSpace(d.Team.ValueString())
	if team == "" {
		return "", "", "", fmt.Errorf("`team` is required")
	}

	count := 0
	if !d.Player.IsNull() && !d.Player.IsUnknown() {
		if v := strings.TrimSpace(d.Player.ValueString()); v != "" {
			count++
			kind = "player"
			value = v
		}
	}
	if !d.Selector.IsNull() && !d.Selector.IsUnknown() {
		if v := strings.TrimSpace(d.Selector.ValueString()); v != "" {
			count++
			kind = "selector"
			value = v
		}
	}
	if !d.EntityID.IsNull() && !d.EntityID.IsUnknown() {
		if v := strings.TrimSpace(d.EntityID.ValueString()); v != "" {
			count++
			kind = "entity"
			value = v
		}
	}

	if count == 0 {
		return "", "", "", fmt.Errorf("exactly one of `player`, `selector`, or `entity_id` must be set")
	}
	if count > 1 {
		return "", "", "", fmt.Errorf("only one of `player`, `selector`, or `entity_id` may be set")
	}
	return team, kind, value, nil
}

func pickTeamMemberTarget(d teamMemberResourceModel) (kind, value string) {
	if !d.Player.IsNull() && !d.Player.IsUnknown() {
		if v := strings.TrimSpace(d.Player.ValueString()); v != "" {
			return "player", v
		}
	}
	if !d.Selector.IsNull() && !d.Selector.IsUnknown() {
		if v := strings.TrimSpace(d.Selector.ValueString()); v != "" {
			return "selector", v
		}
	}
	if !d.EntityID.IsNull() && !d.EntityID.IsUnknown() {
		if v := strings.TrimSpace(d.EntityID.ValueString()); v != "" {
			return "entity", v
		}
	}
	return "", ""
}

func parseTeamMemberID(id string) (kind, value string) {
	parts := strings.SplitN(id, "|", 3)
	if len(parts) == 3 {
		return parts[1], parts[2]
	}
	return "", ""
}

