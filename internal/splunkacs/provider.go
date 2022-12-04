package splunkacs

import (
	"context"
	"os"

	"github.com/atanaspam/splunkacs-api-go/splunkacs"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &AcsProvider{}
var _ provider.ProviderWithMetadata = &AcsProvider{}

type AcsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type AcsProviderModel struct {
	DeploymentName types.String `tfsdk:"deployment_name"`
	AuthToken      types.String `tfsdk:"token"`
}

func (p *AcsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "splunkacs"
	resp.Version = p.version
}

func (p *AcsProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "The Splunk Admin Config Service (ACS) provider can interact with the resources supported by the Splunk Admin Config Service. The provider needs to be configured with the proper credentials before it can be used. It requires terraform version 1.0 or later.",
		Attributes: map[string]tfsdk.Attribute{
			"deployment_name": {
				Type:                types.StringType,
				Optional:            true,
				MarkdownDescription: "he URL prefix of your Splunk Cloud Platform deployment (e.g. csms-2io6tw-47150). Can be set via the `SPLUNK_DEPLOYMENT_NAME` environment variable.",
			},
			"token": {
				Type:                types.StringType,
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The JWT authentication token you create in Splunk Cloud Platform. Can be set via the `SPLUNK_AUTH_TOKEN` environment variable.",
			},
		},
	}, nil
}

func (p *AcsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AcsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.DeploymentName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("deployment_name"),
			"Unknown Splunk Deployment Name",
			"The provider cannot create the Splunk Admin Config API client as there is an unknown configuration value for the Deployment Name. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SPLUNK_DEPLOYMENT_NAME environment variable.",
		)
	}

	if data.AuthToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Splunk Authentication Token Value",
			"The provider cannot create the Splunk Admin Config API client as there is an unknown configuration value for the Token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SPLUNK_AUTH_TOKEN environment variable.",
		)
	}

	deployment_name := os.Getenv("SPLUNK_DEPLOYMENT_NAME")
	token := os.Getenv("SPLUNK_AUTH_TOKEN")

	if !data.DeploymentName.IsNull() {
		deployment_name = data.DeploymentName.ValueString()
	}

	if !data.AuthToken.IsNull() {
		token = data.AuthToken.ValueString()
	}

	if deployment_name == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("deployment_name"),
			"Missing Splunk Deployment Name",
			"The provider cannot create the Splunk Admin Config API client as there is a missing or empty value for the Splunk Deployment Name. "+
				"Set the deployment_name value in the configuration or use the SPLUNK_DEPLOYMENT_NAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Splunk Authentication Token Value",
			"The provider cannot create the Splunk Admin Config API client as there is a missing or empty value for the Splunk Authentication Token. "+
				"Set the token value in the configuration or use the SPLUNK_AUTH_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := splunkacs.NewClient(deployment_name, token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Splunk Admin Config API Client",
			"An unexpected error occurred when creating the Splunk Admin Config API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"HashiCups Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AcsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHecTokenResource,
	}
}

func (p *AcsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHecTokenDataSource,
	}
}

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &AcsProvider{
		version: "dev",
	}
}
