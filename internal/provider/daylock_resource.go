package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ resource.Resource = (*daylockResource)(nil)
var _ resource.ResourceWithImportState = (*daylockResource)(nil)

type daylockResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

type daylockResource struct {
	client *minecraft.Client
}

func NewDaylockResource() resource.Resource {
	return &daylockResource{}
}

func (r *daylockResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_daylock"
}

func (r *daylockResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Locks or unlocks the world time to permanent day on a Minecraft Java server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource ID. Always `\"default\"` for this global server setting.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Set to `true` to lock the world at daytime; `false` to restore the normal day/night cycle.",
				Required:            true,
			},
		},
	}
}

func (r *daylockResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *daylockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data daylockResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Enabled.ValueBool() {
		if err := r.client.EnableDayLock(ctx); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to enable daylock: %s", err))
			return
		}
	} else {
		if err := r.client.DisableDayLock(ctx); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to disable daylock: %s", err))
			return
		}
	}

	data.ID = types.StringValue("default")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *daylockResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data daylockResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *daylockResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data daylockResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Enabled.ValueBool() {
		if err := r.client.EnableDayLock(ctx); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to enable daylock: %s", err))
			return
		}
	} else {
		if err := r.client.DisableDayLock(ctx); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to disable daylock: %s", err))
			return
		}
	}

	if data.ID.IsNull() || data.ID.IsUnknown() {
		data.ID = types.StringValue("default")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *daylockResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	if err := r.client.DisableDayLock(ctx); err != nil {
		resp.Diagnostics.AddWarning("Delete Warning", fmt.Sprintf("Failed to disable daylock during destroy: %s", err))
	}
}

func (r *daylockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "default" {
		resp.Diagnostics.AddError("Import Error", "Expected import ID to be \"default\" for the global daylock setting.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "default")...)
}

