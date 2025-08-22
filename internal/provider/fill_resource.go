package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = fillResourceType{}
var _ tfsdk.Resource = fillResource{}
var _ tfsdk.ResourceWithImportState = fillResource{}

type fillResourceType struct{}

func (t fillResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Fill a **cuboid region** with a single block material (wraps `/fill`).",

		Attributes: map[string]tfsdk.Attribute{
			"material": {
				MarkdownDescription: "Block ID to fill with (e.g. `minecraft:stone`).",
				Required:            true,
				Type:                types.StringType,
				// Material can be changed in-place via /fill on Update, so no ForceNew.
			},

			"start": {
				MarkdownDescription: "Inclusive start corner of the cuboid.",
				Required:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"x": {
						MarkdownDescription: "X coordinate.",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(), // position changes => new resource
						},
					},
					"y": {
						MarkdownDescription: "Y coordinate.",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"z": {
						MarkdownDescription: "Z coordinate.",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
				}),
			},

			"end": {
				MarkdownDescription: "Inclusive end corner of the cuboid.",
				Required:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"x": {
						MarkdownDescription: "X coordinate.",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"y": {
						MarkdownDescription: "Y coordinate.",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
					"z": {
						MarkdownDescription: "Z coordinate.",
						Type:                types.NumberType,
						Required:            true,
						PlanModifiers: tfsdk.AttributePlanModifiers{
							tfsdk.RequiresReplace(),
						},
					},
				}),
			},

			"id": {
				Computed:            true,
				Type:                types.StringType,
				MarkdownDescription: "Terraform ID for this filled region.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (t fillResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return fillResource{provider: provider}, diags
}

type fillResourceData struct {
	Id       types.String `tfsdk:"id"`
	Material string       `tfsdk:"material"`
	Start    struct {
		X int `tfsdk:"x"`
		Y int `tfsdk:"y"`
		Z int `tfsdk:"z"`
	} `tfsdk:"start"`
	End struct {
		X int `tfsdk:"x"`
		Y int `tfsdk:"y"`
		Z int `tfsdk:"z"`
	} `tfsdk:"end"`
}

type fillResource struct {
	provider provider
}

func (r fillResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data fillResourceData
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

	if err := client.FillBlock(ctx,
		data.Material,
		data.Start.X, data.Start.Y, data.Start.Z,
		data.End.X, data.End.Y, data.End.Z,
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill region: %s", err))
		return
	}

	data.Id = types.String{Value: fmt.Sprintf(
		"%s|%d,%d,%d->%d,%d,%d",
		data.Material,
		data.Start.X, data.Start.Y, data.Start.Z,
		data.End.X, data.End.Y, data.End.Z,
	)}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r fillResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// No drift detection yet; keep state as-is.
	var data fillResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r fillResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Only material is mutable; coordinates are ForceNew.
	var data fillResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	if err := client.FillBlock(ctx,
		data.Material,
		data.Start.X, data.Start.Y, data.Start.Z,
		data.End.X, data.End.Y, data.End.Z,
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update filled region: %s", err))
		return
	}

	// ID stays the same unless you want it to include material.
	// If you prefer material-agnostic ID, comment the next line out.
	data.Id = types.String{Value: fmt.Sprintf(
		"%s|%d,%d,%d->%d,%d,%d",
		data.Material,
		data.Start.X, data.Start.Y, data.Start.Z,
		data.End.X, data.End.Y, data.End.Z,
	)}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r fillResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data fillResourceData
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

	if err := client.FillBlock(ctx,
		"minecraft:air",
		data.Start.X, data.Start.Y, data.Start.Z,
		data.End.X, data.End.Y, data.End.Z,
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to clear region: %s", err))
		return
	}
}

func (r fillResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Import by ID string. Caller must supply matching config (material/start/end) in HCL.
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
