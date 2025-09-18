package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ tfsdk.Provider = &provider{}

type provider struct {
	address  string
	password string

	configured bool
	version    string
}

type providerData struct {
	Address  types.String `tfsdk:"address"`
	Password types.String `tfsdk:"password"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var address string
	if data.Address.Null {
		address = os.Getenv("MINECRAFT_ADDRESS")
	} else {
		address = data.Address.Value
	}

	if address == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Address cannot be an empty string",
		)
		return
	}

	var password string
	if data.Password.Null {
		password = os.Getenv("MINECRAFT_PASSWORD")
	} else {
		password = data.Password.Value
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Password cannot be an empty string",
		)
		return
	}

	p.address = address
	p.password = password
	p.configured = true
}

func (p *provider) GetClient(ctx context.Context) (*minecraft.Client, error) {
	client, err := minecraft.New(p.address, p.password)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"minecraft_block":       blockResourceType{},
		"minecraft_entity":      entityResourceType{},
		"minecraft_bed":         bedResourceType{},
		"minecraft_stairs":      stairsResourceType{},
		"minecraft_chest":       chestResourceType{},
		"minecraft_team":        teamResourceType{},
		"minecraft_team_member": teamMemberResourceType{},
		"minecraft_fill":        fillResourceType{},
		"minecraft_gamerule":    gameruleResourceType{},
		"minecraft_op": 		 opResourceType{},
	}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"address": {
				MarkdownDescription: "The RCON address of the Minecraft server",
				Required:            true,
				Type:                types.StringType,
			},
			"password": {
				MarkdownDescription: "The RCON address of the Minecraft server",
				Required:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}
