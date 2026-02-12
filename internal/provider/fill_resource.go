package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ resource.Resource = (*fillResource)(nil)
var _ resource.ResourceWithImportState = (*fillResource)(nil)

type fillResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Material types.String `tfsdk:"material"`
	Start    struct {
		X types.Int32 `tfsdk:"x"`
		Y types.Int32 `tfsdk:"y"`
		Z types.Int32 `tfsdk:"z"`
	} `tfsdk:"start"`
	End struct {
		X types.Int32 `tfsdk:"x"`
		Y types.Int32 `tfsdk:"y"`
		Z types.Int32 `tfsdk:"z"`
	} `tfsdk:"end"`
}

type fillResource struct {
	client *minecraft.Client
}

func NewFillResource() resource.Resource {
	return &fillResource{}
}

func (r *fillResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fill"
}

func (r *fillResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fill a cuboid region with a single block material (wraps `/fill`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Terraform ID for this filled region.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"material": schema.StringAttribute{
				MarkdownDescription: "Block ID to fill with (e.g. `minecraft:stone`).",
				Required:            true,
			},
			"start": schema.SingleNestedAttribute{
				MarkdownDescription: "Inclusive start corner of the cuboid.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x": schema.Int32Attribute{
						MarkdownDescription: "X coordinate.",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"y": schema.Int32Attribute{
						MarkdownDescription: "Y coordinate.",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"z": schema.Int32Attribute{
						MarkdownDescription: "Z coordinate.",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"end": schema.SingleNestedAttribute{
				MarkdownDescription: "Inclusive end corner of the cuboid.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"x": schema.Int32Attribute{
						MarkdownDescription: "X coordinate.",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"y": schema.Int32Attribute{
						MarkdownDescription: "Y coordinate.",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
					"z": schema.Int32Attribute{
						MarkdownDescription: "Z coordinate.",
						Required:            true,
						PlanModifiers: []planmodifier.Int32{
							int32planmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

func (r *fillResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *fillResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data fillResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.FillBlock(
		ctx,
		data.Material.ValueString(),
		int(data.Start.X.ValueInt32()), int(data.Start.Y.ValueInt32()), int(data.Start.Z.ValueInt32()),
		int(data.End.X.ValueInt32()), int(data.End.Y.ValueInt32()), int(data.End.Z.ValueInt32()),
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fill region: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf(
		"%s|%d,%d,%d->%d,%d,%d",
		data.Material.ValueString(),
		data.Start.X.ValueInt32(), data.Start.Y.ValueInt32(), data.Start.Z.ValueInt32(),
		data.End.X.ValueInt32(), data.End.Y.ValueInt32(), data.End.Z.ValueInt32(),
	))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fillResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data fillResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fillResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data fillResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.FillBlock(
		ctx,
		data.Material.ValueString(),
		int(data.Start.X.ValueInt32()), int(data.Start.Y.ValueInt32()), int(data.Start.Z.ValueInt32()),
		int(data.End.X.ValueInt32()), int(data.End.Y.ValueInt32()), int(data.End.Z.ValueInt32()),
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update filled region: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf(
		"%s|%d,%d,%d->%d,%d,%d",
		data.Material.ValueString(),
		data.Start.X.ValueInt32(), data.Start.Y.ValueInt32(), data.Start.Z.ValueInt32(),
		data.End.X.ValueInt32(), data.End.Y.ValueInt32(), data.End.Z.ValueInt32(),
	))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fillResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data fillResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.FillBlock(
		ctx,
		"minecraft:air",
		int(data.Start.X.ValueInt32()), int(data.Start.Y.ValueInt32()), int(data.Start.Z.ValueInt32()),
		int(data.End.X.ValueInt32()), int(data.End.Y.ValueInt32()), int(data.End.Z.ValueInt32()),
	); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to clear region: %s", err))
		return
	}
}

func (r *fillResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

