package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicraft/terraform-provider-minecraft/internal/minecraft"
)

var _ resource.Resource = (*gameruleResource)(nil)
var _ resource.ResourceWithImportState = (*gameruleResource)(nil)

type gameruleResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type gameruleResource struct {
	client *minecraft.Client
}

func NewGameruleResource() resource.Resource {
	return &gameruleResource{}
}

func (r *gameruleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gamerule"
}

func (r *gameruleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a Minecraft gamerule. `value` is a string: use `true`/`false` for boolean rules, or an integer for numeric rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource ID (same as `name`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Gamerule key (e.g., `keepInventory`, `doDaylightCycle`, `randomTickSpeed`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value as string: `true`/`false` for boolean rules, or an integer for numeric rules.",
				Required:            true,
			},
		},
	}
}

func (r *gameruleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gameruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data gameruleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(data.Name.ValueString())
	val := strings.TrimSpace(data.Value.ValueString())

	if i, convErr := strconv.Atoi(val); convErr == nil {
		if err := r.client.SetGameRuleInt(ctx, name, i); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set gamerule %q: %s", name, err))
			return
		}
	} else {
		lv := strings.ToLower(val)
		if lv == "true" || lv == "false" {
			if err := r.client.SetGameRuleBool(ctx, name, lv == "true"); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set gamerule %q: %s", name, err))
				return
			}
		} else {
			resp.Diagnostics.AddError("Invalid Gamerule Value", fmt.Sprintf("Value %q is neither an integer nor true/false.", val))
			return
		}
	}

	data.ID = types.StringValue(name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gameruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data gameruleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(data.Name.ValueString())
	raw, err := r.client.GetGameRule(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read gamerule %q: %s", name, err))
		return
	}

	data.Value = types.StringValue(strings.TrimSpace(raw))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gameruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data gameruleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(data.Name.ValueString())
	val := strings.TrimSpace(data.Value.ValueString())

	if i, convErr := strconv.Atoi(val); convErr == nil {
		if err := r.client.SetGameRuleInt(ctx, name, i); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set gamerule %q: %s", name, err))
			return
		}
	} else {
		lv := strings.ToLower(val)
		if lv == "true" || lv == "false" {
			if err := r.client.SetGameRuleBool(ctx, name, lv == "true"); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set gamerule %q: %s", name, err))
				return
			}
		} else {
			resp.Diagnostics.AddError("Invalid Gamerule Value", fmt.Sprintf("Value %q is neither an integer nor true/false.", val))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gameruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	var data gameruleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(data.Name.ValueString())
	if err := r.client.ResetGameRuleToDefault(ctx, name); err != nil {
		resp.Diagnostics.AddWarning("Reset Warning", fmt.Sprintf("Could not reset gamerule %q to default: %s", name, err))
	}
}

func (r *gameruleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			"Provider client is not configured. Please configure the provider before using this resource.",
		)
		return
	}

	name := strings.TrimSpace(req.ID)
	raw, err := r.client.GetGameRule(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to read gamerule %q: %s", name, err))
		return
	}

	data := gameruleResourceModel{
		ID:    types.StringValue(name),
		Name:  types.StringValue(name),
		Value: types.StringValue(strings.TrimSpace(raw)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

