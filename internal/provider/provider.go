package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/netbirdio/terraform-provider-netbird/internal/sdk"
)

var _ provider.Provider = (*netbirdProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &netbirdProvider{}
	}
}

type netbirdProvider struct {
}

// NetbirdProviderModel describes the provider data model.
type NetbirdProviderModel struct {
	ServerURL        types.String `tfsdk:"server_url"`
	TokenAuth        types.String `tfsdk:"token_auth"`
	OauthCredentials types.String `tfsdk:"oauth_credentials"`
	OauthIssuer      types.String `tfsdk:"oauth_issuer"`
}

func (p *netbirdProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `NetBird REST API: API to manipulate groups, rules, policies and retrieve information about peers and users`,
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				MarkdownDescription: "Server URL (defaults to https://api.netbird.io)",
				Optional:            true,
				Required:            false,
			},
			"token_auth": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"oauth_credentials": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"oauth_issuer": schema.StringAttribute{
				Optional:  true,
				Sensitive: false,
			},
		},
	}
}

func (p *netbirdProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data NetbirdProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serverURL := data.ServerURL.ValueString()
	if serverURL == "" {
		serverURL = "https://api.netbird.io"
	}
	var tokenAuthString string
	if !data.TokenAuth.IsNull() {
		tokenAuthString = "Token " + data.TokenAuth.ValueString()
	}
	if !data.OauthCredentials.IsNull() && data.OauthCredentials.ValueString() != "" {
		bearerToken, _diags := getTokenUsingOauth(data.OauthCredentials.ValueString(), data.OauthIssuer.ValueString())
		if _diags != nil {
			resp.Diagnostics.Append(_diags...)
			return
		}
		tokenAuthString = "Bearer " + bearerToken
	}
	addRequestAuth := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", tokenAuthString)
		return nil
	}
	client, err := sdk.NewClientWithResponses(serverURL, sdk.WithRequestEditorFn(addRequestAuth))
	if err != nil {
		resp.Diagnostics.AddError("failed to create client", err.Error())
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func getTokenUsingOauth(oauthCredentialsJsonFilePath string, isserUrlString string) (string, diag.Diagnostics) {
	var token string
	var _diags diag.Diagnostics
	oauthCredentialsJson, err := os.ReadFile(oauthCredentialsJsonFilePath) //read the content of file
	if err != nil {
		_diags.AddError("Failed to read file", err.Error())
		return "", _diags
	}
	var tokenRequest TokenRequest
	err = json.Unmarshal(oauthCredentialsJson, &tokenRequest)
	if err != nil {
		_diags.AddError("Failed to unmarshal request body", err.Error())
		return "", _diags
	}
	formBody := url.Values{}
	formBody.Set("client_id", tokenRequest.ClientID)
	formBody.Set("client_secret", tokenRequest.ClientSecret)
	formBody.Set("grant_type", tokenRequest.GrantType)
	formBody.Set("scope", tokenRequest.Scope)
	req, err := http.NewRequest(http.MethodPost, isserUrlString, strings.NewReader(formBody.Encode()))
	if err != nil {
		_diags.AddError("Failed to create request", err.Error())
		return "", _diags
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	httpClient := http.Client{}

	res, err := httpClient.Do(req)
	if err != nil {
		_diags.AddError("error when call oauth bearer token request", err.Error())
		return "", _diags
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		_diags.AddError("error when reading bearer token resonse", err.Error())
		return "", _diags
	}
	var tokenResponse TokenResponse
	err = json.Unmarshal(respBody, &tokenResponse)
	if err != nil {
		_diags.AddError("error when unmarshalling bearer token reponse", err.Error())
		return "", _diags
	}
	token = tokenResponse.AccessToken
	return token, _diags
}

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (p *netbirdProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netbird"
}

func (p *netbirdProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGroupsDataSource,
		NewRouteDataSource,
	}
}

func (p *netbirdProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSetupKeyResource,
		NewGroupResource,
		NewRouteResource,
	}
}
