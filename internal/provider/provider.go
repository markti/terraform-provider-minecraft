package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ provider.Provider = &minecraftProvider{}

type minecraftProvider struct {
	address  string
	password string

	configured bool
	version    string
}

type minecraftProviderModel struct {
	Address  types.String `tfsdk:"address"`
	Password types.String `tfsdk:"password"`
}

func (p *minecraftProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "minecraft"
	resp.Version = p.version
}

func (p *minecraftProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				MarkdownDescription: "The RCON address of the Minecraft server",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The RCON password of the Minecraft server",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *minecraftProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data minecraftProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var address string
	if data.Address.IsNull() {
		address = os.Getenv("MINECRAFT_ADDRESS")
	} else {
		address = data.Address.ValueString()
	}

	if address == "" {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Address cannot be an empty string",
		)
		return
	}

	var password string
	if data.Password.IsNull() {
		password = os.Getenv("MINECRAFT_PASSWORD")
	} else {
		password = data.Password.ValueString()
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

	if os.Getenv("MINECRAFT_SKIP_CONNECT") == "true" {
		return
	}

	client, err := minecraft.New(p.address, p.password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			fmt.Sprintf("An unexpected error was encountered trying to connect to the Minecraft server: %s", err.Error()),
		)
		return
	}

	resp.ResourceData = client
}

func (p *minecraftProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBlockResource,
		NewEntityResource,
		NewBedResource,
		NewStairsResource,
		NewChestResource,
		NewTeamResource,
		NewTeamMemberResource,
		NewFillResource,
		NewGameruleResource,
		NewOpResource,
		NewGamemodeResource,
		NewDaylockResource,
		NewSheepResource,
		NewZombieResource,
	}
}

func (p *minecraftProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &minecraftProvider{
			version: version,
		}
	}
}
