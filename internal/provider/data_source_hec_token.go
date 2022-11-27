package provider

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

func (d *hecTokenDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Fetches the details about an individual HEC Token.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Id of the HEC token.",
				Type:                types.StringType,
				Computed:            true,
			},
			"allowed_indexes": {
				MarkdownDescription: "The indexes a HEC Token is allowed to publish it.",
				Type:                types.SetType{ElemType: types.StringType},
				Computed:            true,
			},
			"default_host": {
				MarkdownDescription: "The default host associated with a HEC Token.",
				Type:                types.StringType,
				Computed:            true,
			},
			"default_index": {
				MarkdownDescription: "The default index associated with a HEC Token.",
				Type:                types.StringType,
				Computed:            true,
			},
			"default_source": {
				MarkdownDescription: "The default source value assigned to the data from this HEC Token.",
				Type:                types.StringType,
				Computed:            true,
			},
			"default_sourcetype": {
				MarkdownDescription: "The default sourcetype assigned to the data from this HEC Token.",
				Type:                types.StringType,
				Computed:            true,
			},
			"disabled": {
				MarkdownDescription: "The state of the HEC token.",
				Type:                types.BoolType,
				Computed:            true,
			},
			"name": {
				MarkdownDescription: "The name of the HEC token.",
				Type:                types.StringType,
				Required:            true,
			},
			"use_ack": {
				MarkdownDescription: "Is indexer acknoldegment enabled for this HEC token.",
				Type:                types.BoolType,
				Computed:            true,
			},
			"token": {
				MarkdownDescription: "The token value.",
				Type:                types.StringType,
				Computed:            true,
			},
		},
	}, nil
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
