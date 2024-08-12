package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netbirdio/terraform-provider-netbird/internal/sdk"
)

var _ datasource.DataSource = (*groupsDataSource)(nil)

func NewGroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

type groupsDataSourceModel struct {
	Groups []groupsModel `tfsdk:"groups"`
}

type groupsModel struct {
	Id    types.String     `tfsdk:"id"`
	Name  types.String     `tfsdk:"name"`
	Peers []peersTypeModel `tfsdk:"peers"`
}

// coffeesIngredientsModel maps coffee ingredients data
type peersTypeModel struct {
	Id types.String `tfsdk:"id"`
}

type groupsDataSource struct {
	client *sdk.ClientWithResponses
}

func (d *groupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *groupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "The unique identifier of a group",
							MarkdownDescription: "The unique identifier of a group",
						},
						"name": schema.StringAttribute{
							Required:            true,
							Description:         "Group name identifier",
							MarkdownDescription: "Group name identifier",
						},
						"peers": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *groupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = client
}

func (d *groupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupsDataSourceModel

	res, err := d.client.GetApiGroupsWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failure to invoke get groups API", err.Error())
		return
	}

	for _, group := range *res.JSON200 {
		groupsState := groupsModel{
			Id:   types.StringValue(group.Id),
			Name: types.StringValue(group.Name),
		}

		for _, peer := range group.Peers {
			groupsState.Peers = append(groupsState.Peers, peersTypeModel{
				Id: types.StringValue(peer.Id),
			})
		}
		state.Groups = append(state.Groups, groupsState)

	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
