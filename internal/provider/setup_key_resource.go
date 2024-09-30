package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/netbirdio/terraform-provider-netbird/internal/provider/resource_setup_key"
	"github.com/netbirdio/terraform-provider-netbird/internal/sdk"
)

var _ resource.Resource = (*setupKeyResource)(nil)

func NewSetupKeyResource() resource.Resource {
	return &setupKeyResource{}
}

type setupKeyResource struct {
	client *sdk.ClientWithResponses
}

func (r *setupKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setup_key"
}

func (r *setupKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_setup_key.SetupKeyResourceSchema(ctx)
}

func (r *setupKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sdk.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *setupKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_setup_key.SetupKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.PostApiSetupKeysWithResponse(ctx, toCreateSetupKeyApiRequest(data))
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke create setup key API", err.Error())
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from create setupkey API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}

	createdSetupKey, diags := toSetupKeyModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	createdSetupKey.ExpiresIn = data.ExpiresIn
	createdSetupKey.RotationId = data.RotationId

	resp.Diagnostics.Append(resp.State.Set(ctx, &createdSetupKey)...)
}

func (r *setupKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_setup_key.SetupKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetApiSetupKeysKeyIdWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke get setup key API", err.Error())
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from read setupkey API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}

	setupKey, diags := toSetupKeyModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	setupKey.ExpiresIn = data.ExpiresIn
	setupKey.RotationId = data.RotationId
	resp.Diagnostics.Append(resp.State.Set(ctx, &setupKey)...)
}

func (r *setupKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state resource_setup_key.SetupKeyModel
	var plan resource_setup_key.SetupKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.PutApiSetupKeysKeyIdWithResponse(ctx, state.Id.ValueString(), toSetupKeyApiRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke update setup key API", err.Error())
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from update setupkey API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}

	setupKey, diags := toSetupKeyModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	setupKey.ExpiresIn = plan.ExpiresIn
	setupKey.RotationId = plan.RotationId
	resp.Diagnostics.Append(resp.State.Set(ctx, &setupKey)...)
}

func (r *setupKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_setup_key.SetupKeyModel
	var diags diag.Diagnostics

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Revoked = types.BoolValue(true)
	data.AutoGroups, diags = types.ListValueFrom(ctx, types.StringType, []string{})
	if diags.HasError() {
		return
	}
	res, err := r.client.PutApiSetupKeysKeyIdWithResponse(ctx, data.Id.ValueString(), toSetupKeyApiRequest(data))
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke update setup key API", err.Error())
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from delete setupkey API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}
}

func toCreateSetupKeyApiRequest(data resource_setup_key.SetupKeyModel) sdk.CreateSetupKeyRequest {
	autoGroups := make([]string, len(data.AutoGroups.Elements()))
	for i, v := range data.AutoGroups.Elements() {
		if !v.IsUnknown() && !v.IsNull() {
			value, ok := v.(types.String)
			if ok {
				autoGroups[i] = value.ValueString()
			}
		}
	}

	name := ""
	if !data.Name.IsUnknown() && !data.Name.IsNull() {
		name = data.Name.ValueString()
	}

	ephemeral := new(bool)
	if !data.Ephemeral.IsUnknown() && !data.Ephemeral.IsNull() {
		ephemeral = data.Ephemeral.ValueBoolPointer()
	}

	expiresIn := int64(0)
	if !data.ExpiresIn.IsUnknown() && !data.ExpiresIn.IsNull() {
		expiresIn = data.ExpiresIn.ValueInt64()
	}

	keyType := ""
	if !data.Type.IsUnknown() && !data.Type.IsNull() {
		keyType = data.Type.ValueString()
	}

	usageLimit := int64(0)
	if !data.UsageLimit.IsUnknown() && !data.UsageLimit.IsNull() {
		usageLimit = data.UsageLimit.ValueInt64()
	}

	return sdk.CreateSetupKeyRequest{
		AutoGroups: autoGroups,
		Ephemeral:  ephemeral,
		ExpiresIn:  int(expiresIn),
		Name:       name,
		Type:       keyType,
		UsageLimit: int(usageLimit),
	}
}

func toSetupKeyApiRequest(data resource_setup_key.SetupKeyModel) sdk.SetupKeyRequest {
	autoGroups := make([]string, len(data.AutoGroups.Elements()))
	for i, v := range data.AutoGroups.Elements() {
		if !v.IsUnknown() && !v.IsNull() {
			value, ok := v.(types.String)
			if ok {
				autoGroups[i] = value.ValueString()
			}
		}
	}

	name := ""
	if !data.Name.IsUnknown() && !data.Name.IsNull() {
		name = data.Name.ValueString()
	}

	ephemeral := new(bool)
	if !data.Ephemeral.IsUnknown() && !data.Ephemeral.IsNull() {
		ephemeral = data.Ephemeral.ValueBoolPointer()
	}

	expiresIn := int64(0)
	if !data.ExpiresIn.IsUnknown() && !data.ExpiresIn.IsNull() {
		expiresIn = data.ExpiresIn.ValueInt64()
	}

	keyType := ""
	if !data.Type.IsUnknown() && !data.Type.IsNull() {
		keyType = data.Type.ValueString()
	}

	usageLimit := int64(0)
	if !data.UsageLimit.IsUnknown() && !data.UsageLimit.IsNull() {
		usageLimit = data.UsageLimit.ValueInt64()
	}

	revoked := false
	if !data.Revoked.IsUnknown() && !data.Revoked.IsNull() {
		revoked = data.Revoked.ValueBool()
	}

	return sdk.SetupKeyRequest{
		AutoGroups: autoGroups,
		Ephemeral:  ephemeral,
		ExpiresIn:  int(expiresIn),
		Name:       name,

		Type:       keyType,
		UsageLimit: int(usageLimit),
		Revoked:    revoked,
	}
}

func toSetupKeyModel(ctx context.Context, data *sdk.SetupKey) (resource_setup_key.SetupKeyModel, diag.Diagnostics) {
	model := resource_setup_key.SetupKeyModel{
		Ephemeral:  types.BoolValue(data.Ephemeral),
		Expires:    types.StringValue(data.Expires.String()),
		Id:         types.StringValue(data.Id),
		Key:        types.StringValue(data.Key),
		LastUsed:   types.StringValue(data.LastUsed.String()),
		Name:       types.StringValue(data.Name),
		Revoked:    types.BoolValue(data.Revoked),
		State:      types.StringValue(data.State),
		Type:       types.StringValue(data.Type),
		UpdatedAt:  types.StringValue(data.UpdatedAt.String()),
		UsageLimit: types.Int64Value(int64(data.UsageLimit)),
		UsedTimes:  types.Int64Value(int64(data.UsedTimes)),
		Valid:      types.BoolValue(data.Valid),
	}

	var diags diag.Diagnostics
	var _diags diag.Diagnostics
	var autoGroupsToApply types.List
	if data.AutoGroups != nil {
		autoGroups := make([]string, len(data.AutoGroups))
		copy(autoGroups, data.AutoGroups)
		autoGroupsToApply, _diags = types.ListValueFrom(ctx, types.StringType, autoGroups)
	} else {
		autoGroupsToApply, _diags = types.ListValueFrom(ctx, types.StringType, []string{})
	}
	diags.Append(_diags...)
	model.AutoGroups = autoGroupsToApply

	return model, diags
}
