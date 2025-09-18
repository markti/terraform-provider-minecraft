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
var _ tfsdk.ResourceType = opResourceType{}
var _ tfsdk.Resource = opResource{}
var _ tfsdk.ResourceWithImportState = opResource{}

// -------- Resource Type --------

type opResourceType struct{}

func (t opResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Grants or revokes Minecraft server operator (op) status for a player.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Resource ID (same as `player`).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"player": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "Minecraft player username to grant operator privileges to.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // changing player => ForceNew
				},
			},
		},
	}, nil
}

func (t opResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return opResource{provider: p}, diags
}

// -------- Data & Resource --------

type opResourceData struct {
	ID     types.String `tfsdk:"id"`
	Player types.String `tfsdk:"player"`
}

type opResource struct {
	provider provider
}

// Define the minimal client surface we need (helps with testing/mocking)
type opClient interface {
	CreateOp(ctx context.Context, name string) error
	RemoveOp(ctx context.Context, name string) error
}

// -------- CRUD --------

func (r opResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan opResourceData
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

	player := strings.TrimSpace(plan.Player.Value)
	if player == "" {
		resp.Diagnostics.AddError("Validation Error", "Attribute `player` cannot be empty or whitespace.")
		return
	}

	// Grant op
	if err := client.CreateOp(ctx, player); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to grant operator to %q: %s", player, err))
		return
	}

	plan.ID = types.String{Value: player}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r opResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// No straightforward, portable RCON query to verify op list in this minimal version.
	// Keep state as-is; drift detection can be added later if you expose an API to list ops.
	var state opResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r opResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// No updatable attributes; `player` is ForceNew. Just keep plan as state.
	var plan opResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r opResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state opResourceData
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

	player := strings.TrimSpace(state.Player.Value)
	if player == "" {
		// Nothing to do
		return
	}

	if err := client.RemoveOp(ctx, player); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to revoke operator from %q: %s", player, err))
		return
	}
}

func (r opResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Allow `terraform import minecraft_op.this <playerName>`
	// Set both id and player based on provided ID.
	player := strings.TrimSpace(req.ID)
	if player == "" {
		resp.Diagnostics.AddError("Import Error", "Expected non-empty player name as import ID.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), player)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("player"), player)...)
}
