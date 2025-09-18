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
var _ tfsdk.ResourceType = gamemodeResourceType{}
var _ tfsdk.Resource = gamemodeResource{}
var _ tfsdk.ResourceWithImportState = gamemodeResource{}

// ---------- Resource Type ----------

type gamemodeResourceType struct{}

func (t gamemodeResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Set the default server gamemode or a specific player's gamemode.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Resource ID (`default` or `player:<name>`).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"mode": {
				Type:     types.StringType,
				Required: true,
				MarkdownDescription: "Target gamemode. One of `survival`, `creative`, `adventure`, `spectator`.",
			},
			"player": {
				Type:     types.StringType,
				Optional: true,
				MarkdownDescription: "If set, applies the mode to this player; otherwise sets the server default.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // switching target identity => ForceNew
				},
			},
			"previous_mode": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Best-effort snapshot of the prior mode at create/update time. Used for revert.",
			},
		},
	}, nil
}

func (t gamemodeResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return gamemodeResource{provider: p}, diags
}

// ---------- Data & Resource ----------

type gamemodeResourceData struct {
	ID           types.String `tfsdk:"id"`
	Mode         types.String `tfsdk:"mode"`
	Player       types.String `tfsdk:"player"`
	PreviousMode types.String `tfsdk:"previous_mode"`
}

type gamemodeResource struct {
	provider provider
}

// Minimal client surface we need
type gamemodeClient interface {
	SetDefaultGameMode(ctx context.Context, gamemode string) error
	SetUserGameMode(ctx context.Context, gamemode string, name string) error

	// NEW: explicit getters so we can snapshot previous values
	GetDefaultGameMode(ctx context.Context) (string, error)
	GetUserGameMode(ctx context.Context, name string) (string, error)
}

// ---------- CRUD ----------

func (r gamemodeResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan gamemodeResourceData
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

	mode := strings.ToLower(strings.TrimSpace(plan.Mode.Value))
	if err := validateMode(mode); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	var id string
	var prev string

	player := strings.TrimSpace(plan.Player.Value)
	if player == "" {
		id = "default"

		// Snapshot previous default (best effort)
		if got, e := client.GetDefaultGameMode(ctx); e == nil && got != "" {
			prev = got
		}

		if err := client.SetDefaultGameMode(ctx, mode); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set default gamemode to %q: %s", mode, err))
			return
		}
	} else {
		id = "player:" + player

		// Snapshot previous player mode (best effort)
		if got, e := client.GetUserGameMode(ctx, player); e == nil && got != "" {
			prev = got
		}

		if err := client.SetUserGameMode(ctx, mode, player); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set %q gamemode to %q: %s", player, mode, err))
			return
		}
	}

	plan.ID = types.String{Value: id}
	plan.PreviousMode = types.String{Value: prev}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r gamemodeResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Keep state as-is. (Optional future enhancement: detect drift via getters)
	var state gamemodeResourceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r gamemodeResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state gamemodeResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	mode := strings.ToLower(strings.TrimSpace(plan.Mode.Value))
	if err := validateMode(mode); err != nil {
		resp.Diagnostics.AddError("Validation Error", err.Error())
		return
	}

	player := strings.TrimSpace(plan.Player.Value)
	if player == "" {
		// Refresh previous_mode for default (best effort)
		prev := state.PreviousMode.Value
		if got, e := client.GetDefaultGameMode(ctx); e == nil && got != "" {
			prev = got
		}
		plan.PreviousMode = types.String{Value: prev}

		if err := client.SetDefaultGameMode(ctx, mode); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set default gamemode to %q: %s", mode, err))
			return
		}
	} else {
		// Refresh previous_mode for player (best effort)
		prev := state.PreviousMode.Value
		if got, e := client.GetUserGameMode(ctx, player); e == nil && got != "" {
			prev = got
		}
		plan.PreviousMode = types.String{Value: prev}

		if err := client.SetUserGameMode(ctx, mode, player); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set %q gamemode to %q: %s", player, mode, err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r gamemodeResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state gamemodeResourceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	// Revert if requested and we know a previous value
	prev := strings.TrimSpace(state.PreviousMode.Value)
	player := strings.TrimSpace(state.Player.Value)

	if prev != "" {
		if player == "" {
			if err := client.SetDefaultGameMode(ctx, prev); err != nil {
				resp.Diagnostics.AddWarning("Restore Warning", fmt.Sprintf("Failed to restore default gamemode to %q: %s", prev, err))
			}
		} else {
			if err := client.SetUserGameMode(ctx, prev, player); err != nil {
				resp.Diagnostics.AddWarning("Restore Warning", fmt.Sprintf("Failed to restore %q gamemode to %q: %s", player, prev, err))
			}
		}
	}

	// Nothing else to delete remotely; resource is imperative.
}

func (r gamemodeResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Accept "default" or "player:<name>"
	id := strings.TrimSpace(req.ID)
	if id == "" {
		resp.Diagnostics.AddError("Import Error", "Expected `default` or `player:<name>` as import ID.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), id)...)

	if id == "default" {
		// user must set desired mode in config
		return
	}

	if strings.HasPrefix(id, "player:") {
		player := strings.TrimPrefix(id, "player:")
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("player"), player)...)
		return
	}

	resp.Diagnostics.AddError("Import Error", "Unrecognized import ID. Use `default` or `player:<name>`.")
}

// ---------- Helpers ----------

func validateMode(m string) error {
	switch m {
	case "survival", "creative", "adventure", "spectator":
		return nil
	default:
		return fmt.Errorf("mode must be one of: survival, creative, adventure, spectator (got %q)", m)
	}
}
