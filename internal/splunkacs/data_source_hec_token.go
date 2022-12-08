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
var _ datasource.DataSource = &hecTokenDataSource{}
var _ datasource.DataSourceWithConfigure = &hecTokenDataSource{}

func NewHecTokenDataSource() datasource.DataSource {
	return &hecTokenDataSource{}
}

// HecTokenDataSource defines the data source implementation.
type hecTokenDataSource struct {
	client *splunkacs.SplunkAcsClient
}

// HttpEventCollectorToken maps the HttpEventCollectorToken schema data
type HttpEventCollectorToken struct {
	Id                types.String   `tfsdk:"id"`
	AllowedIndexes    []types.String `tfsdk:"allowed_indexes"`
	DefaultHost       types.String   `tfsdk:"default_host"`
	DefaultIndex      types.String   `tfsdk:"default_index"`
	DefaultSource     types.String   `tfsdk:"default_source"`
	DefaultSourcetype types.String   `tfsdk:"default_sourcetype"`
	Disabled          types.Bool     `tfsdk:"disabled"`
	Name              types.String   `tfsdk:"name"`
	UseACK            types.Bool     `tfsdk:"use_ack"`
	Token             types.String   `tfsdk:"token"`
}

func (d *hecTokenDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hec_token"
}

func (d *hecTokenDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the details about an individual HEC Token.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the HEC token.",
				Computed:            true,
			},
			"allowed_indexes": schema.SetAttribute{
				MarkdownDescription: "The indexes the HEC Token is allowed to publish data to.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"default_host": schema.StringAttribute{
				MarkdownDescription: "The default Splunk host associated with th HEC Token.",
				Computed:            true,
			},
			"default_index": schema.StringAttribute{
				MarkdownDescription: "The default index associated with the HEC Token.",
				Computed:            true,
			},
			"default_source": schema.StringAttribute{
				MarkdownDescription: "The default source value assigned to the data from the HEC Token.",
				Computed:            true,
			},
			"default_sourcetype": schema.StringAttribute{
				MarkdownDescription: "The default sourcetype assigned to the data from the HEC Token.",
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "The state of the HEC token.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the HEC token.",
				Computed:            true,
				Required:            true,
			},
			"use_ack": schema.BoolAttribute{
				MarkdownDescription: "Is indexer acknoldegment enabled for the HEC token.",
				Computed:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The token value.",
				Computed:            true,
			},
		},
	}
}

func (d *hecTokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *hecTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HttpEventCollectorToken

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hecResp, _, err := d.client.GetHecToken(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get HEC token", err.Error())
		return
	}

	for _, index := range hecResp.HttpEventCollector.Spec.AllowedIndexes {
		state.AllowedIndexes = append(state.AllowedIndexes, types.StringValue(index))
	}
	state.DefaultHost = types.StringValue(hecResp.HttpEventCollector.Spec.DefaultHost)
	state.DefaultIndex = types.StringValue(hecResp.HttpEventCollector.Spec.DefaultIndex)
	state.DefaultSource = types.StringValue(hecResp.HttpEventCollector.Spec.DefaultSource)
	state.DefaultSourcetype = types.StringValue(hecResp.HttpEventCollector.Spec.DefaultSourcetype)
	state.Disabled = types.BoolValue(hecResp.HttpEventCollector.Spec.Disabled)
	state.Name = types.StringValue(hecResp.HttpEventCollector.Spec.Name)
	state.UseACK = types.BoolValue(hecResp.HttpEventCollector.Spec.UseACK)
	state.Token = types.StringValue(hecResp.HttpEventCollector.Token)
	state.Id = types.StringValue(hecResp.HttpEventCollector.Spec.Name)

	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Failed to set state for data source")
		return
	}
}
