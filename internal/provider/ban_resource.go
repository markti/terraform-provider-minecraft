package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type banResourceType struct{}

func (r banResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Unique ID for this ban resource.",
			},
			"player": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "Player to ban.",
			},
			"reason": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "Reason for ban.",
			},
		},
	}, nil
}

func (r banResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return banResourceImpl{provider: p.(*provider)}, nil
}

type banResourceImpl struct {
	provider *provider
}

func (r banResourceImpl) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data struct {
		ID     types.String `tfsdk:"id"`
		Player types.String `tfsdk:"player"`
		Reason types.String `tfsdk:"reason"`
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Minecraft client", err.Error())
		return
	}

	err = client.BanPlayer(ctx, data.Player.Value, data.Reason.Value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to ban player", err.Error())
		return
	}

	data.ID = data.Player // Use player name as unique ID

	resp.State.Set(ctx, &data)
}

func (r banResourceImpl) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
}
func (r banResourceImpl) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data struct {
		ID     types.String `tfsdk:"id"`
		Player types.String `tfsdk:"player"`
		Reason types.String `tfsdk:"reason"`
	}
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Minecraft client", err.Error())
		return
	}

	err = client.BanPlayer(ctx, data.Player.Value, data.Reason.Value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to ban player", err.Error())
		return
	}

	data.ID = data.Player

	resp.State.Set(ctx, &data)
}
func (r banResourceImpl) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data struct {
		ID     types.String `tfsdk:"id"`
		Player types.String `tfsdk:"player"`
		Reason types.String `tfsdk:"reason"`
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Minecraft client", err.Error())
		return
	}

	err = client.UnbanPlayer(ctx, data.Player.Value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unban player", err.Error())
		return
	}
}
