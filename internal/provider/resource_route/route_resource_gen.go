// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package resource_route

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func RouteResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Required:            true,
				Description:         "Route description",
				MarkdownDescription: "Route description",
			},
			"domains": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Domain list to be dynamically resolved. Conflicts with network",
				MarkdownDescription: "Domain list to be dynamically resolved. Conflicts with network",
			},
			"enabled": schema.BoolAttribute{
				Required:            true,
				Description:         "Route status",
				MarkdownDescription: "Route status",
			},
			"groups": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				Description:         "Group IDs containing routing peers",
				MarkdownDescription: "Group IDs containing routing peers",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The unique identifier of a route",
				MarkdownDescription: "The unique identifier of a route",
			},
			"keep_route": schema.BoolAttribute{
				Required:            true,
				Description:         "Indicate if the route should be kept after a domain doesn't resolve that IP anymore",
				MarkdownDescription: "Indicate if the route should be kept after a domain doesn't resolve that IP anymore",
			},
			"masquerade": schema.BoolAttribute{
				Required:            true,
				Description:         "Indicate if peer should masquerade traffic to this route's prefix",
				MarkdownDescription: "Indicate if peer should masquerade traffic to this route's prefix",
			},
			"metric": schema.Int64Attribute{
				Required:            true,
				Description:         "Route metric number. Lowest number has higher priority",
				MarkdownDescription: "Route metric number. Lowest number has higher priority",
				Validators: []validator.Int64{
					int64validator.Between(1, 9999),
				},
			},
			"network": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Network range in CIDR format, Conflicts with domains",
				MarkdownDescription: "Network range in CIDR format, Conflicts with domains",
			},
			"network_id": schema.StringAttribute{
				Required:            true,
				Description:         "Route network identifier, to group HA routes",
				MarkdownDescription: "Route network identifier, to group HA routes",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 40),
				},
			},
			"peer": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Peer Identifier associated with route. This property can not be set together with `peer_groups`",
				MarkdownDescription: "Peer Identifier associated with route. This property can not be set together with `peer_groups`",
			},
			"peer_groups": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Peers Group Identifier associated with route. This property can not be set together with `peer`",
				MarkdownDescription: "Peers Group Identifier associated with route. This property can not be set together with `peer`",
			},
		},
	}
}

type RouteModel struct {
	Description types.String `tfsdk:"description"`
	Domains     types.List   `tfsdk:"domains"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Groups      types.List   `tfsdk:"groups"`
	Id          types.String `tfsdk:"id"`
	KeepRoute   types.Bool   `tfsdk:"keep_route"`
	Masquerade  types.Bool   `tfsdk:"masquerade"`
	Metric      types.Int64  `tfsdk:"metric"`
	Network     types.String `tfsdk:"network"`
	NetworkId   types.String `tfsdk:"network_id"`
	Peer        types.String `tfsdk:"peer"`
	PeerGroups  types.List   `tfsdk:"peer_groups"`
}
