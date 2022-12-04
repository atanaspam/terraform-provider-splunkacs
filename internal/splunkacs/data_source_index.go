package splunkacs

import (
	"context"
	"fmt"

	"github.com/atanaspam/splunkacs-api-go/splunkacs"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &hecTokenDataSource{}
var _ datasource.DataSourceWithConfigure = &hecTokenDataSource{}

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

func (d *indexDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Fetches the details about an individual Index.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of the Index.",
				Type:                types.StringType,
				Computed:            true,
			},
			"name": {
				MarkdownDescription: "The name of the Index.",
				Type:                types.StringType,
				Required:            true,
			},
			"data_type": {
				MarkdownDescription: "The type of data the index holds. Possible values: `event` or `metric`.",
				Type:                types.StringType,
				Computed:            true,
				// 	Validators: []tfsdk.AttributeValidator{
				// 		stringvalidator.OneOf(validator.AllowedIndexTypes()...),
				// 	},
			},
			"searchable_days": {
				MarkdownDescription: "Number of days the index is searchable..",
				Type:                types.Int64Type,
				Computed:            true,
			},
			"max_data_size_mb": {
				MarkdownDescription: "The maximum size of the index in megabytes.",
				Type:                types.Int64Type,
				Computed:            true,
			},
			"total_event_count": {
				MarkdownDescription: "The total number of events in an index.",
				Type:                types.StringType,
				Computed:            true,
			},
			"total_raw_size_mb": {
				MarkdownDescription: "The total amount of raw data in an index in megabytes.",
				Type:                types.StringType,
				Computed:            true,
			},
		},
	}, nil
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
