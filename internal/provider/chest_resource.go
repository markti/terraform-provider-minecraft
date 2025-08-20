package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tfsdk.ResourceType = chestResourceType{}
var _ tfsdk.Resource = chestResource{}
var _ tfsdk.ResourceWithImportState = chestResource{}

type chestResourceType struct{}

func (t chestResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A Minecraft chest. Can be a single chest or a double chest (two blocks side by side).",
		Attributes: map[string]tfsdk.Attribute{
			"position": {
				MarkdownDescription: "The position of the first chest block.",
				Required:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"x": {
						Type:     types.NumberType,
						Required: true,
					},
					"y": {
						Type:     types.NumberType,
						Required: true,
					},
					"z": {
						Type:     types.NumberType,
						Required: true,
					},
				}),
			},
			"size": {
				MarkdownDescription: "The chest size: `single` or `double`.",
				Required:            true,
				Type:                types.StringType,
			},
			"trapped": {
				MarkdownDescription: "Whether this is a trapped chest. Defaults to false.",
				Optional:            true,
				Type:                types.BoolType,
			},
			"waterlogged": {
				MarkdownDescription: "Whether the chest is waterlogged. Defaults to false.",
				Optional:            true,
				Type:                types.BoolType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "ID of the chest resource.",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t chestResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return chestResource{provider: provider}, diags
}

type chestResourceData struct {
	Id          types.String `tfsdk:"id"`
	Size        string       `tfsdk:"size"`
	Trapped     *bool        `tfsdk:"trapped"`
	Waterlogged *bool        `tfsdk:"waterlogged"`
	Position    struct {
		X int `tfsdk:"x"`
		Y int `tfsdk:"y"`
		Z int `tfsdk:"z"`
	} `tfsdk:"position"`
}

type chestResource struct {
	provider provider
}

func (r chestResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data chestResourceData
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

	waterlogged := false
	if data.Waterlogged != nil {
		waterlogged = *data.Waterlogged
	}

	trapped := false
	if data.Trapped != nil {
		trapped = *data.Trapped
	}

	material := "minecraft:chest"
	if trapped {
		material = "minecraft:trapped_chest"
	}

	switch data.Size {
	case "single":
		block := fmt.Sprintf(`%s[type=single,waterlogged=%t]`, material, waterlogged)
		err = client.CreateBlock(ctx, block, data.Position.X, data.Position.Y, data.Position.Z)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place single chest: %s", err))
			return
		}
	case "double":
		blockLeft := fmt.Sprintf(`%s[type=left,waterlogged=%t]`, material, waterlogged)
		blockRight := fmt.Sprintf(`%s[type=right,waterlogged=%t]`, material, waterlogged)
		err = client.CreateBlock(ctx, blockLeft, data.Position.X, data.Position.Y, data.Position.Z)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place left half of double chest: %s", err))
			return
		}
		err = client.CreateBlock(ctx, blockRight, data.Position.X+1, data.Position.Y, data.Position.Z)
		if err != nil {
			_ = client.DeleteBlock(ctx, data.Position.X, data.Position.Y, data.Position.Z)
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to place right half of double chest: %s", err))
			return
		}
	default:
		resp.Diagnostics.AddError("Validation Error", "size must be 'single' or 'double'")
		return
	}

	data.Id = types.String{Value: fmt.Sprintf("chest-%d-%d-%d", data.Position.X, data.Position.Y, data.Position.Z)}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r chestResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data chestResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r chestResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data chestResourceData
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

	waterlogged := false
	if data.Waterlogged != nil {
		waterlogged = *data.Waterlogged
	}

	trapped := false
	if data.Trapped != nil {
		trapped = *data.Trapped
	}

	material := "minecraft:chest"
	if trapped {
		material = "minecraft:trapped_chest"
	}

	switch data.Size {
	case "single":
		block := fmt.Sprintf(`%s[type=single,waterlogged=%t]`, material, waterlogged)
		err = client.CreateBlock(ctx, block, data.Position.X, data.Position.Y, data.Position.Z)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update single chest: %s", err))
			return
		}
	case "double":
		blockLeft := fmt.Sprintf(`%s[type=left,waterlogged=%t]`, material, waterlogged)
		blockRight := fmt.Sprintf(`%s[type=right,waterlogged=%t]`, material, waterlogged)
		err = client.CreateBlock(ctx, blockLeft, data.Position.X, data.Position.Y, data.Position.Z)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update left half of double chest: %s", err))
			return
		}
		err = client.CreateBlock(ctx, blockRight, data.Position.X+1, data.Position.Y, data.Position.Z)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update right half of double chest: %s", err))
			return
		}
	default:
		resp.Diagnostics.AddError("Validation Error", "size must be 'single' or 'double'")
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r chestResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data chestResourceData
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

	_ = client.DeleteBlock(ctx, data.Position.X, data.Position.Y, data.Position.Z)
	if data.Size == "double" {
		_ = client.DeleteBlock(ctx, data.Position.X+1, data.Position.Y, data.Position.Z)
	}
}

func (r chestResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
