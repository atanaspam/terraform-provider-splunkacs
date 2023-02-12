package splunkacs

import (
	"context"
	"fmt"

	"github.com/atanaspam/splunkacs-api-go/splunkacs"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &indexDataSource{}
var _ datasource.DataSourceWithConfigure = &indexDataSource{}

func NewStackStatusDataSource() datasource.DataSource {
	return &stackStatusDataSource{}
}

// stackDataSource defines the data source implementation.
type stackStatusDataSource struct {
	client *splunkacs.SplunkAcsClient
}

type stackStatusSchema struct {
	Id      types.String `tfsdk:"id"`
	Type    types.String `tfsdk:"type"`
	Version types.String `tfsdk:"version"`
}

func (d *stackStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_status"
}

func (d *stackStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the version of the current Splunk stack.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Index.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The stack type.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The current version of the stack.",
				Computed:            true,
			},
		},
	}
}

func (d *stackStatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*splunkacs.SplunkAcsClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *splunkacs.SplunkAcsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *stackStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state stackStatusSchema

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	stackStatusResp, _, err := d.client.GetStackStatus()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Stack Status during data source read", err.Error())
		return
	}

	state.Type = types.StringValue(stackStatusResp.Infrastructure.StackType)
	state.Version = types.StringValue(stackStatusResp.Infrastructure.StackVersion)

	state.Id = types.StringValue(d.client.Url)

	tflog.Trace(ctx, "read a stack_status data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state for data source")
		return
	}
}
