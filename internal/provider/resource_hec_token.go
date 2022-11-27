package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/atanaspam/splunkacs-api-go/splunkacs"
	// "github.com/hashicorp/terraform-plugin-framework-timeouts/timeouts"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &HecTokenResource{}
var _ resource.ResourceWithImportState = &HecTokenResource{}

func NewHecTokenResource() resource.Resource {
	return &HecTokenResource{}
}

// HecTokenResource defines the resource implementation.
type HecTokenResource struct {
	client *splunkacs.SplunkAcsClient
}

// HecTokenResourceModel describes the resource data model.
type HecTokenResourceModel struct {
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
	// Timeouts          types.Object   `tfsdk:"timeouts"`
}

func (r *HecTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hec_token"
}

func (r *HecTokenResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Creates a Http Event Collector Token",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Id of the HEC token.",
				Type:                types.StringType,
				Computed:            true,
			},
			"allowed_indexes": {
				MarkdownDescription: "The indexes a HEC Token is allowed to publish it.",
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"default_host": {
				MarkdownDescription: "The default host associated with a HEC Token.",
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"default_index": {
				MarkdownDescription: "The default index associated with a HEC Token.",
				Type:                types.StringType,
				Optional:            false,
				Required:            true,
			},
			"default_source": {
				MarkdownDescription: "The default source value assigned to the data from this HEC Token.",
				Type:                types.StringType,
				Optional:            true,
			},
			"default_sourcetype": {
				MarkdownDescription: "The default sourcetype assigned to the data from this HEC Token.",
				Type:                types.StringType,
				Optional:            true,
			},
			"disabled": {
				MarkdownDescription: "The state of the HEC token.",
				Type:                types.BoolType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				MarkdownDescription: "The name of the HEC token.",
				Type:                types.StringType,
				Required:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"use_ack": {
				MarkdownDescription: "Is indexer acknoldegment enabled for this HEC token.",
				Type:                types.BoolType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"token": {
				MarkdownDescription: "The token value.",
				Type:                types.StringType,
				Computed:            true,
				Optional:            true,
			},
			// "timeouts": timeouts.Attributes(ctx, timeouts.Opts{
			// 	Create: true,
			// 	Update: true,
			// }),
		},
	}, nil
}

func (r *HecTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*splunkacs.SplunkAcsClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *splunkacs.SplunkAcsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *HecTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *HecTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare AllowedIndexes
	allowedIndexes := make([]string, 0)
	for _, index := range data.AllowedIndexes {
		allowedIndexes = append(allowedIndexes, index.ValueString())
	}

	hecToken := splunkacs.HecTokenSpec{
		AllowedIndexes:    allowedIndexes,
		DefaultHost:       data.DefaultHost.ValueString(),
		DefaultIndex:      data.DefaultIndex.ValueString(),
		DefaultSource:     data.DefaultSource.ValueString(),
		DefaultSourcetype: data.DefaultSourcetype.ValueString(),
		Disabled:          data.Disabled.ValueBool(),
		Name:              data.Name.ValueString(),
		UseACK:            data.UseACK.ValueBool(),
	}

	request := splunkacs.HttpEventCollectorCreateRequest{HecTokenSpec: hecToken}

	// Set and initiate the timeout
	// defaultCreateTimeout := 2 * time.Minute
	// createTimeout := timeouts.Create(ctx, data.Timeouts, defaultCreateTimeout)
	// ctx, cancel := context.WithTimeout(ctx, createTimeout)
	// defer cancel()

	hecResp, _, err := r.client.CreateHecToken(request)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while creating HEC Token", err.Error())
		return
	}

	hecGetResp, err := waitHecCreatePropagation(ctx, r.client, hecResp)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while waiting for HEC Token", err.Error())
		return
	}

	allowedIndexesResult := make([]types.String, 0)
	for _, index := range hecGetResp.HttpEventCollector.Spec.AllowedIndexes {
		allowedIndexesResult = append(allowedIndexesResult, types.StringValue(index))
	}
	data.AllowedIndexes = allowedIndexesResult
	data.DefaultHost = types.StringValue(hecGetResp.HttpEventCollector.Spec.DefaultHost)
	data.DefaultIndex = types.StringValue(hecGetResp.HttpEventCollector.Spec.DefaultIndex)
	data.DefaultSource = types.StringValue(hecGetResp.HttpEventCollector.Spec.DefaultSource)
	data.DefaultSourcetype = types.StringValue(hecGetResp.HttpEventCollector.Spec.DefaultSourcetype)
	data.Disabled = types.BoolValue(hecGetResp.HttpEventCollector.Spec.Disabled)
	data.Name = types.StringValue(hecGetResp.HttpEventCollector.Spec.Name)
	data.UseACK = types.BoolValue(hecGetResp.HttpEventCollector.Spec.UseACK)
	data.Token = types.StringValue(hecGetResp.HttpEventCollector.Token)
	data.Id = types.StringValue(hecGetResp.HttpEventCollector.Spec.Name)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HecTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *HecTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hecResp, _, err := r.client.GetHecToken(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read HEC token", err.Error())
		return
	}

	allowedIndexesResult := make([]types.String, 0)
	for _, index := range hecResp.HttpEventCollector.Spec.AllowedIndexes {
		allowedIndexesResult = append(allowedIndexesResult, types.StringValue(index))
	}
	data.AllowedIndexes = allowedIndexesResult
	data.DefaultHost = types.StringValue(hecResp.HttpEventCollector.Spec.DefaultHost)
	data.DefaultIndex = types.StringValue(hecResp.HttpEventCollector.Spec.DefaultIndex)
	data.DefaultSource = types.StringValue(hecResp.HttpEventCollector.Spec.DefaultSource)
	data.DefaultSourcetype = types.StringValue(hecResp.HttpEventCollector.Spec.DefaultSourcetype)
	data.Disabled = types.BoolValue(hecResp.HttpEventCollector.Spec.Disabled)
	data.Name = types.StringValue(hecResp.HttpEventCollector.Spec.Name)
	data.UseACK = types.BoolValue(hecResp.HttpEventCollector.Spec.UseACK)
	data.Token = types.StringValue(hecResp.HttpEventCollector.Token)
	data.Id = types.StringValue(hecResp.HttpEventCollector.Spec.Name)

	tflog.Trace(ctx, "read a resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HecTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *HecTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare AllowedIndexes
	allowedIndexes := make([]string, 0)
	for _, index := range data.AllowedIndexes {
		allowedIndexes = append(allowedIndexes, index.ValueString())
	}

	hecToken := splunkacs.HecTokenSpec{
		AllowedIndexes:    allowedIndexes,
		DefaultHost:       data.DefaultHost.ValueString(),
		DefaultIndex:      data.DefaultIndex.ValueString(),
		DefaultSource:     data.DefaultSource.ValueString(),
		DefaultSourcetype: data.DefaultSourcetype.ValueString(),
		Disabled:          data.Disabled.ValueBool(),
		Name:              data.Name.ValueString(),
		UseACK:            data.UseACK.ValueBool(),
	}

	request := splunkacs.HttpEventCollectorUpdateRequest{HecTokenSpec: hecToken}

	_, _, err := r.client.UpdateHecToken(data.Name.ValueString(), request)
	// hecUpdateResp, _, err := r.client.UpdateHecToken(data.Name.ValueString(), request)
	// Splunk Docs and Splunk API response seem to differ. While the snippet below makes sense, it is commented out
	// because the Splunk API actually does not return the code.
	// if err != nil || hecUpdateResp.Code != "202" {
	// 	resp.Diagnostics.AddError("Unexpected error while updating HEC Token", err.Error())
	// 	return
	// }
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while updating HEC Token", err.Error())
		return
	}

	// Given the response from the Splunk API, we need further API calls to confirm if the changes have taken effect.
	hecGetResp, err := waitHecUpdatePropagation(ctx, r.client, hecToken)
	if err != nil {
		resp.Diagnostics.AddError("Encountered an error while waiting for HEC Token update to propagate", err.Error())
		return
	}

	allowedIndexesResult := make([]types.String, 0)
	for _, index := range hecGetResp.HttpEventCollector.Spec.AllowedIndexes {
		allowedIndexesResult = append(allowedIndexesResult, types.StringValue(index))
	}
	data.AllowedIndexes = allowedIndexesResult
	data.DefaultHost = types.StringValue(hecGetResp.HttpEventCollector.Spec.DefaultHost)
	data.DefaultIndex = types.StringValue(hecGetResp.HttpEventCollector.Spec.DefaultIndex)
	data.DefaultSource = types.StringValue(hecGetResp.HttpEventCollector.Spec.DefaultSource)
	data.DefaultSourcetype = types.StringValue(hecGetResp.HttpEventCollector.Spec.DefaultSourcetype)
	data.Disabled = types.BoolValue(hecGetResp.HttpEventCollector.Spec.Disabled)
	data.Name = types.StringValue(hecGetResp.HttpEventCollector.Spec.Name)
	data.UseACK = types.BoolValue(hecGetResp.HttpEventCollector.Spec.UseACK)
	data.Token = types.StringValue(hecGetResp.HttpEventCollector.Token)
	data.Id = types.StringValue(hecGetResp.HttpEventCollector.Spec.Name)

	tflog.Trace(ctx, "updated a resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HecTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *HecTokenResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, _, err := r.client.DeleteHecToken(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while updating HEC Token", err.Error())
		return
	}
}

func (r *HecTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

/* HELPERS */
func waitHecCreatePropagation(ctx context.Context, client *splunkacs.SplunkAcsClient, hecCreateResponse *splunkacs.HttpEventCollectorCreateResponse) (*splunkacs.HttpEventCollectorGetResponse, error) {
	// TODO: Get rid of the for loop. Technically the timeouts should cover for us and we can fo a while true
	// TODO: Add logging inside for each iteration in the loop
	// TODO: How do I do this using the native framework? Seems to be possible in SDKv2...
	i := 0
	retries := 20
	for i < retries {
		tflog.Debug(ctx, fmt.Sprintf("waiting for HEC token to become available. Retry: %d", i))
		hecResp, httpResp, err := client.GetHecToken(hecCreateResponse.CreateResponseItem.Spec.Name)
		if err != nil && httpResp.StatusCode != 404 {
			tflog.Error(ctx, "encountered an unexpected error while waiting for HEC to become avaialable")
			return nil, err
		} else if err != nil && httpResp.StatusCode == 404 {
			i++
			time.Sleep(10 * time.Second)
			continue
		}
		return hecResp, nil
	}
	return nil, fmt.Errorf("failed to fetch a valid HEC token defintion after %d retries", retries)
}

// TODO can we pass an interface instead of the specific spec. This will allow us to make this waiter generic
// TODO why doesn't this exist in the plugin framework? :(
// https://github.com/hashicorp/terraform-plugin-framework/issues/513
// Reads the state of a HEC token and compares it against an expected state until a timeout is reached, hoping to work around eventual consistency
func waitHecUpdatePropagation(ctx context.Context, client *splunkacs.SplunkAcsClient, expectedState splunkacs.HecTokenSpec) (*splunkacs.HttpEventCollectorGetResponse, error) {
	i := 0
	retries := 10
	var lastResp *splunkacs.HttpEventCollectorGetResponse
	for i < retries {
		tflog.Debug(ctx, fmt.Sprintf("waiting for HEC token to become eventually consistent. Retry: %d", i))
		hecResp, _, err := client.GetHecToken(expectedState.Name)
		if err != nil {
			tflog.Error(ctx, "encountered an unexpected error while waiting for HEC token propagation")
			return nil, err
		}
		if hecResp.HttpEventCollector.Spec.Equal(expectedState) {
			return hecResp, nil
		}
		lastResp = hecResp
		i++
		time.Sleep(10 * time.Second)
		continue
	}
	tflog.Error(ctx, fmt.Sprintf("%v", lastResp.HttpEventCollector))
	return nil, fmt.Errorf("failed to obtain the expected HEC token values after %d retries", i)
}
