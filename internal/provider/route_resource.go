package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netbirdio/terraform-provider-netbird/internal/provider/resource_route"
	"github.com/netbirdio/terraform-provider-netbird/internal/sdk"
)

var _ resource.Resource = (*routeResource)(nil)

func NewRouteResource() resource.Resource {
	return &routeResource{}
}

type routeResource struct {
	client *sdk.ClientWithResponses
}

func (r *routeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

func (r *routeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_route.RouteResourceSchema(ctx)
}

func (r *routeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *routeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_route.RouteModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	routeRequest := toCreateRouteApiRequest(data)
	res, err := r.client.PostApiRoutesWithResponse(ctx, routeRequest)
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke create route API", err.Error())
		return
	}
	createRoute, diags := toRouteModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &createRoute)...)

}

func (r *routeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_route.RouteModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetApiRoutesRouteIdWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke get route API", err.Error())
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from read route API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}
	route, diags := toRouteModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &route)...)
}

func (r *routeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state resource_route.RouteModel
	var plan resource_route.RouteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.PutApiRoutesRouteIdWithResponse(ctx, state.Id.ValueString(), toCreateRouteApiRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke update route API", err.Error())
		return
	}

	route, diags := toRouteModel(ctx, res.JSON200)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from update route API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &route)...)
}

func (r *routeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_route.RouteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.DeleteApiRoutesRouteIdWithResponse(ctx, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("failure to invoke delete of route API", err.Error())
		return
	}

	if res.StatusCode() != 200 {
		resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from delete route API. Got an unexpected response code %d", res.StatusCode()), string(res.Body))
		return
	}

}

func toRouteModel(ctx context.Context, data *sdk.Route) (resource_route.RouteModel, diag.Diagnostics) {

	model := resource_route.RouteModel{
		Description: types.StringValue(data.Description),
		Enabled:     types.BoolValue(data.Enabled),
		KeepRoute:   types.BoolValue(data.KeepRoute),
		Masquerade:  types.BoolValue(data.Masquerade),
		Metric:      types.Int64Value(int64(data.Metric)),
		Network:     types.StringValue(*data.Network),
		NetworkId:   types.StringValue(data.NetworkId),
		Peer:        types.StringValue(*data.Peer),
		Id:          types.StringValue(data.Id),
	}

	var groups types.List
	var peerGroups types.List
	var domains types.List
	var _diags diag.Diagnostics
	var diags diag.Diagnostics
	if data.Groups != nil {
		_groups := make([]string, len(data.Groups))
		copy(_groups, data.Groups)
		groups, _diags = types.ListValueFrom(ctx, types.StringType, _groups)
	} else {
		groups, _diags = types.ListValueFrom(ctx, types.StringType, []string{})
	}
	diags.Append(_diags...)

	if data.Domains != nil {
		_domains := make([]string, len(*data.Domains))
		copy(_domains, *data.Domains)
		domains, diags = types.ListValueFrom(ctx, types.StringType, _domains)
	} else {
		domains, diags = types.ListValueFrom(ctx, types.StringType, []string{})
	}
	diags.Append(_diags...)

	if data.PeerGroups != nil {
		_peerGroups := make([]string, len(*data.PeerGroups))
		copy(_peerGroups, *data.PeerGroups)
		peerGroups, diags = types.ListValueFrom(ctx, types.StringType, _peerGroups)
	} else {
		peerGroups, diags = types.ListValueFrom(ctx, types.StringType, []string{})
	}
	diags.Append(_diags...)

	model.Groups = groups
	model.Domains = domains
	model.PeerGroups = peerGroups

	return model, diags
}

func toCreateRouteApiRequest(data resource_route.RouteModel) sdk.RouteRequest {
	groups := make([]string, len(data.Groups.Elements()))
	for i, v := range data.Groups.Elements() {
		if !v.IsUnknown() && !v.IsNull() {
			value, ok := v.(types.String)
			if ok {
				groups[i] = value.ValueString()
			}
		}
	}
	var domainsPoint *[]string
	if len(data.Domains.Elements()) > 0 {
		domains := make([]string, len(data.Domains.Elements()))
		for i, v := range data.Domains.Elements() {
			if !v.IsUnknown() && !v.IsNull() {
				value, ok := v.(types.String)
				if ok {
					domains[i] = value.ValueString()
				}
			}
		}
		domainsPoint = &domains
	}
	var peerGroupsPoint *[]string
	if len(data.PeerGroups.Elements()) > 0 {
		peerGroups := make([]string, len(data.PeerGroups.Elements()))
		for i, v := range data.PeerGroups.Elements() {
			if !v.IsUnknown() && !v.IsNull() {
				value, ok := v.(types.String)
				if ok {
					peerGroups[i] = value.ValueString()
				}
			}
			peerGroupsPoint = &peerGroups
		}
	}
	description := ""
	if !data.Description.IsUnknown() && !data.Description.IsNull() {
		description = data.Description.ValueString()
	}

	enabled := false
	if !data.Enabled.IsUnknown() && !data.Enabled.IsNull() {
		enabled = data.Enabled.ValueBool()
	}

	keepRoute := false
	if !data.KeepRoute.IsUnknown() && !data.KeepRoute.IsNull() {
		keepRoute = data.KeepRoute.ValueBool()
	}

	masquerade := false
	if !data.Masquerade.IsUnknown() && !data.Masquerade.IsNull() {
		masquerade = data.Masquerade.ValueBool()
	}

	metric := int(0)
	if !data.Metric.IsUnknown() && !data.Metric.IsNull() {
		metric = int(data.Metric.ValueInt64())
	}

	var netpoint *string
	if !data.Network.IsUnknown() && !data.Network.IsNull() {
		netpoint = data.Network.ValueStringPointer()
	}

	networkId := ""
	if !data.NetworkId.IsUnknown() && !data.NetworkId.IsNull() {
		networkId = data.NetworkId.ValueString()
	}

	var peerPoint *string
	if !data.Peer.IsUnknown() && !data.Peer.IsNull() {
		peerPoint = data.Peer.ValueStringPointer()
	}

	routeRequest := sdk.RouteRequest{
		Description: description,
		Enabled:     enabled,
		KeepRoute:   keepRoute,
		Masquerade:  masquerade,
		Metric:      metric,
		Network:     netpoint,
		NetworkId:   networkId,
		Peer:        peerPoint,
		PeerGroups:  peerGroupsPoint,
		Groups:      groups,
		Domains:     domainsPoint,
	}
	return routeRequest
}
