package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure types satisfy framework interfaces
var _ tfsdk.ResourceType = daylockResourceType{}
var _ tfsdk.Resource = daylockResource{}
var _ tfsdk.ResourceWithImportState = daylockResource{}

// -------- Resource Type --------

type daylockResourceType struct{}

func (t daylockResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Locks or unlocks the world time to permanent day on a Minecraft Java server.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Resource ID. Always `\"default\"` for this global server setting.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"enabled": {
				Type:                types.BoolType,
				Required:            true,
				MarkdownDescription: "Set to `true` to lock the world at daytime; `false` to restore the normal day/night cycle.",
			},
		},
	}, nil
}

func (t daylockResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return daylockResource{provider: p}, diags
}

// -------- Data & Resource --------

type daylockResourceData struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

type daylockResource struct {
	provider provider
}

// Minimal client surface needed (easy to mock in tests)
type daylockClient interface {
	SetDayLock(ctx context.Context, enabled bool) error
}

// -------- CRUD --------

func (r daylockResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan daylockResourceData
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

	// Apply desired state
	if err := client.SetDayLock(ctx, plan.Enabled.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to set daylock to %t: %s", plan.Enabled.Value, err))
		return
	}

	// Single global instance; use a fixed id
	plan.ID = types.String{Value: "default"}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r daylockResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// No read API available yet; keep state as-is.
	var state daylockResourceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r daylockResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan daylockResourceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	// Re-apply desired enabled state
	if err := client.SetDayLock(ctx, plan.Enabled.Value); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to set daylock to %t: %s", plan.Enabled.Value, err))
		return
	}

	// Keep the fixed id
	if plan.ID.Null || plan.ID.Unknown {
		plan.ID = types.String{Value: "default"}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r daylockResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	// On delete, best-effort to restore normal cycle (disable daylock).
	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	if err := client.SetDayLock(ctx, false); err != nil {
		// Non-fatal: resource is being removed from state regardless.
		resp.Diagnostics.AddWarning("Delete Warning", fmt.Sprintf("Failed to disable daylock during destroy: %s", err))
	}
}

func (r daylockResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Allow: terraform import minecraft_daylock.default default
	if req.ID != "default" {
		resp.Diagnostics.AddError("Import Error", "Expected import ID to be \"default\" for the global daylock setting.")
		return
	}

	// Set id; we cannot know actual enabled value without a read API, so leave it as-is/unknown.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), "default")...)
}
