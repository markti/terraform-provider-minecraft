package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure interfaces
var _ tfsdk.ResourceType = gameruleResourceType{}
var _ tfsdk.Resource = gameruleResource{}
var _ tfsdk.ResourceWithImportState = gameruleResource{}

type gameruleResourceType struct{}

func (t gameruleResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Manage a Minecraft **gamerule**. `value` is a string: use `true`/`false` for boolean rules, or an integer for numeric rules.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "Resource ID (same as `name`).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"name": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "Gamerule key (e.g., `keepInventory`, `doDaylightCycle`, `randomTickSpeed`).",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(), // changing rule name => ForceNew
				},
			},
			"value": {
				Type:                types.StringType,
				Required:            true,
				MarkdownDescription: "Value as string: `true`/`false` for boolean rules, or an integer for numeric rules.",
			},
		},
	}, nil
}

func (t gameruleResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return gameruleResource{provider: p}, diags
}

type gameruleResource struct {
	provider provider
}

type gameruleData struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (r gameruleResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan gameruleData
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

	name := strings.TrimSpace(plan.Name.Value)
	val := strings.TrimSpace(plan.Value.Value)

	// Infer rule type from value: int -> SetGameRuleInt, else true/false -> SetGameRuleBool
	if i, convErr := strconv.Atoi(val); convErr == nil {
		if err := client.SetGameRuleInt(ctx, name, i); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set gamerule %q: %s", name, err))
			return
		}
	} else {
		lv := strings.ToLower(val)
		if lv == "true" || lv == "false" {
			if err := client.SetGameRuleBool(ctx, name, lv == "true"); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set gamerule %q: %s", name, err))
				return
			}
		} else {
			resp.Diagnostics.AddError("Invalid Gamerule Value", fmt.Sprintf("Value %q is neither an integer nor true/false.", val))
			return
		}
	}

	plan.ID = types.String{Value: name}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r gameruleResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state gameruleData
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

	name := strings.TrimSpace(state.Name.Value)
	raw, err := client.GetGameRule(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read gamerule %q: %s", name, err))
		return
	}

	state.Value = types.String{Value: strings.TrimSpace(raw)}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r gameruleResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Same as Create
	var plan gameruleData
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

	name := strings.TrimSpace(plan.Name.Value)
	val := strings.TrimSpace(plan.Value.Value)

	if i, convErr := strconv.Atoi(val); convErr == nil {
		if err := client.SetGameRuleInt(ctx, name, i); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set gamerule %q: %s", name, err))
			return
		}
	} else {
		lv := strings.ToLower(val)
		if lv == "true" || lv == "false" {
			if err := client.SetGameRuleBool(ctx, name, lv == "true"); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set gamerule %q: %s", name, err))
				return
			}
		} else {
			resp.Diagnostics.AddError("Invalid Gamerule Value", fmt.Sprintf("Value %q is neither an integer nor true/false.", val))
			return
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r gameruleResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state gameruleData
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

	name := strings.TrimSpace(state.Name.Value)

	// Reset to vanilla default; warn if unknown
	if err := client.ResetGameRuleToDefault(ctx, name); err != nil {
		resp.Diagnostics.AddWarning("Reset Warning", fmt.Sprintf("Could not reset gamerule %q to default: %s", name, err))
	}
}

func (r gameruleResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Import by rule name; read the current value
	name := strings.TrimSpace(req.ID)

	client, err := r.provider.GetClient(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create client: %s", err))
		return
	}

	raw, err := client.GetGameRule(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to read gamerule %q: %s", name, err))
		return
	}

	var st gameruleData
	st.ID = types.String{Value: name}
	st.Name = types.String{Value: name}
	st.Value = types.String{Value: strings.TrimSpace(raw)}

	diags := resp.State.Set(ctx, &st)
	resp.Diagnostics.Append(diags...)
}
