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

func NewIndexDataSource() datasource.DataSource {
	return &indexDataSource{}
}

// indexDataSource defines the data source implementation.
type indexDataSource struct {
	client *splunkacs.SplunkAcsClient
}

func (d *indexDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index"
}

func (d *indexDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the details about an individual Index.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the Index.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Index.",
				Computed:            true,
				Required:            true,
			},
			"data_type": schema.StringAttribute{
				MarkdownDescription: "The type of data the index holds. Possible values: `event` or `metric`.",
				Computed:            true,
			},
			"searchable_days": schema.Int64Attribute{
				MarkdownDescription: "Number of days the index is searchable.",
				Computed:            true,
			},
			"max_data_size_mb": schema.Int64Attribute{
				MarkdownDescription: "The maximum size of the index in megabytes.",
				Computed:            true,
			},
			"total_event_count": schema.StringAttribute{
				MarkdownDescription: "The total number of events in the index.",
				Computed:            true,
			},
			"total_raw_size_mb": schema.StringAttribute{
				MarkdownDescription: "The total amount of raw data in the index in megabytes.",
				Computed:            true,
			},
		},
	}
}

func (d *indexDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *indexDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state Index

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	indexResp, _, err := d.client.GetIndex(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Index during data source read", err.Error())
		return
	}

	state.Name = types.StringValue(indexResp.Name)
	state.DataType = types.StringValue(indexResp.DataType)
	state.SearchableDays = types.Int64Value(int64(indexResp.SearchableDays))
	state.MaxDataSizeMb = types.Int64Value(int64(indexResp.MaxDataSizeMb))
	state.TotalEventCount = types.StringValue(indexResp.TotalEventCount)
	state.TotalRawSizeMb = types.StringValue(indexResp.TotalRawSizeMb)

	state.Id = types.StringValue(indexResp.Name)

	tflog.Trace(ctx, "read an index data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state for data source")
		return
	}
}
