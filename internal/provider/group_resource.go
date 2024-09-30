package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netbirdio/terraform-provider-netbird/internal/provider/resource_group"
	"github.com/netbirdio/terraform-provider-netbird/internal/sdk"
)

var _ resource.Resource = (*groupResource)(nil)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *sdk.ClientWithResponses
}

func (r *groupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_group.GroupResourceSchema(ctx)
}

func (r *groupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_group.GroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	groupRequest := toGroupApiRequest(data)
	res, err := r.client.PostApiGroupsWithResponse(ctx, groupRequest)
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke create groups API", err.Error())
		return
	}
	createGroup, diags := toGroupModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &createGroup)...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_group.GroupModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetApiGroupsGroupIdWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke get groups API", err.Error())
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from read group resource API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}

	group, diags := toGroupModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &group)...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state resource_group.GroupModel
	var plan resource_group.GroupModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.PutApiGroupsGroupIdWithResponse(ctx, state.Id.ValueString(), toGroupApiRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke update group API", err.Error())
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from update group resource API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}

	group, diags := toGroupModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &group)...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_group.GroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.DeleteApiGroupsGroupId(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke delete group API", err.Error())
		return
	}

	if res.StatusCode != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from delete group resource API. Got an unexpected response code %d", res.StatusCode), "")
		return
	}
}

func toGroupApiRequest(data resource_group.GroupModel) sdk.GroupRequest {
	peers := make([]string, len(data.Peers.Elements()))
	for i, v := range data.Peers.Elements() {
		if !v.IsUnknown() && !v.IsNull() {
			value, ok := v.(types.String)
			if ok {
				peers[i] = value.ValueString()
			}
		}
	}

	name := ""
	if !data.Name.IsUnknown() && !data.Name.IsNull() {
		name = data.Name.ValueString()
	}

	return sdk.GroupRequest{
		Name:  name,
		Peers: &peers,
	}

}
func toGroupModel(ctx context.Context, data *sdk.Group) (resource_group.GroupModel, diag.Diagnostics) {

	model := resource_group.GroupModel{
		Name: types.StringValue(data.Name),
		Id:   types.StringValue(data.Id),
	}

	var diags diag.Diagnostics
	var peersToApply types.List
	var _diags diag.Diagnostics

	if data.Peers != nil {
		peers := make([]string, len(data.Peers))
		for i, v := range data.Peers {
			peers[i] = v.Id
		}
		peersToApply, _diags = types.ListValueFrom(ctx, types.StringType, peers)
	} else {
		peersToApply, _diags = types.ListValueFrom(ctx, types.StringType, []string{})
	}
	diags.Append(_diags...)
	model.Peers = peersToApply
	return model, diags
}

// func ListValueOrNull[T any](ctx context.Context, elementType attr.Type, elements []T, diags *diag.Diagnostics) types.List {
// 	if len(elements) == 0 {
// 		return types.ListNull(elementType)
// 	}

// 	result, d := types.ListValueFrom(ctx, elementType, elements)
// 	diags.Append(d...)
// 	return result
// }
