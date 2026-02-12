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

var _ resource.Resource = (*gamemodeResource)(nil)
var _ resource.ResourceWithImportState = (*gamemodeResource)(nil)

type gamemodeResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Mode         types.String `tfsdk:"mode"`
	Player       types.String `tfsdk:"player"`
	PreviousMode types.String `tfsdk:"previous_mode"`
}

type gamemodeResource struct {
	client *minecraft.Client
}

func NewGamemodeResource() resource.Resource {
	return &gamemodeResource{}
}

func (r *gamemodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gamemode"
}

func (r *gamemodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Set the default server gamemode or a specific player's gamemode.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource ID (`default` or `player:<name>`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "Target gamemode. One of `survival`, `creative`, `adventure`, `spectator`.",
				Required:            true,
			},
			"player": schema.StringAttribute{
				MarkdownDescription: "If set, applies the mode to this player; otherwise sets the server default.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"previous_mode": schema.StringAttribute{
				MarkdownDescription: "Best-effort snapshot of the prior mode at create/update time. Used for revert.",
				Computed:            true,
			},
		},
	}
}

func (r *gamemodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gamemodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data gamemodeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mode := strings.ToLower(strings.TrimSpace(data.Mode.ValueString()))
	if err := validateGamemode(mode); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	player := ""
	if !data.Player.IsNull() {
		player = strings.TrimSpace(data.Player.ValueString())
	}

	var id string
	var prev string

	if player == "" {
		id = "default"
		if got, err := r.client.GetDefaultGameMode(ctx); err == nil && got != "" {
			prev = got
		}
		if err := r.client.SetDefaultGameMode(ctx, mode); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set default gamemode to %q: %s", mode, err))
			return
		}
	} else {
		id = "player:" + player
		if got, err := r.client.GetUserGameMode(ctx, player); err == nil && got != "" {
			prev = got
		}
		if err := r.client.SetUserGameMode(ctx, mode, player); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set %q gamemode to %q: %s", player, mode, err))
			return
		}
	}

	data.ID = types.StringValue(id)
	data.PreviousMode = types.StringValue(prev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gamemodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data gamemodeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gamemodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var plan, state gamemodeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mode := strings.ToLower(strings.TrimSpace(plan.Mode.ValueString()))
	if err := validateGamemode(mode); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	player := ""
	if !plan.Player.IsNull() {
		player = strings.TrimSpace(plan.Player.ValueString())
	}

	if player == "" {
		prev := state.PreviousMode.ValueString()
		if got, err := r.client.GetDefaultGameMode(ctx); err == nil && got != "" {
			prev = got
		}
		plan.PreviousMode = types.StringValue(prev)

		if err := r.client.SetDefaultGameMode(ctx, mode); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set default gamemode to %q: %s", mode, err))
			return
		}
	} else {
		prev := state.PreviousMode.ValueString()
		if got, err := r.client.GetUserGameMode(ctx, player); err == nil && got != "" {
			prev = got
		}
		plan.PreviousMode = types.StringValue(prev)

		if err := r.client.SetUserGameMode(ctx, mode, player); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set %q gamemode to %q: %s", player, mode, err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gamemodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var state gamemodeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prev := strings.TrimSpace(state.PreviousMode.ValueString())
	if prev == "" {
		return
	}

	player := strings.TrimSpace(state.Player.ValueString())
	if player == "" {
		if err := r.client.SetDefaultGameMode(ctx, prev); err != nil {
			resp.Diagnostics.AddWarning("Restore Warning", fmt.Sprintf("Failed to restore default gamemode to %q: %s", prev, err))
		}
		return
	}

	if err := r.client.SetUserGameMode(ctx, prev, player); err != nil {
		resp.Diagnostics.AddWarning("Restore Warning", fmt.Sprintf("Failed to restore %q gamemode to %q: %s", player, prev, err))
	}
}

func (r *gamemodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		resp.Diagnostics.AddError("Import Error", "Expected `default` or `player:<name>` as import ID.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	if id == "default" {
		return
	}

	if strings.HasPrefix(id, "player:") {
		player := strings.TrimPrefix(id, "player:")
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("player"), player)...)
		return
	}

	resp.Diagnostics.AddError("Import Error", "Unrecognized import ID. Use `default` or `player:<name>`.")
}

func validateGamemode(m string) error {
	switch m {
	case "survival", "creative", "adventure", "spectator":
		return nil
	default:
		return fmt.Errorf("mode must be one of: survival, creative, adventure, spectator (got %q)", m)
	}
}

