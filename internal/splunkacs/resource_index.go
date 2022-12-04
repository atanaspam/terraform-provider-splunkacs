package splunkacs

import (
	"context"
	"fmt"
	"time"

	"github.com/atanaspam/splunkacs-api-go/splunkacs"
	"github.com/atanaspam/terraform-provider-splunkacs/internal/validator"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &IndexResource{}
var _ resource.ResourceWithImportState = &IndexResource{}

func NewIndexResource() resource.Resource {
	return &IndexResource{}
}

// IndexResource defines the resource implementation.
type IndexResource struct {
	client *splunkacs.SplunkAcsClient
}

// Index maps the Index schema data
type Index struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	DataType        types.String `tfsdk:"data_type"`
	SearchableDays  types.Int64  `tfsdk:"searchable_days"`
	MaxDataSizeMb   types.Int64  `tfsdk:"max_data_size_mb"`
	TotalEventCount types.String `tfsdk:"total_event_count"`
	TotalRawSizeMb  types.String `tfsdk:"total_raw_size_mb"`
}

func (r *IndexResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_index"
}

func (r *IndexResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Creates a Http Event Collector Token",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of the Index.",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				MarkdownDescription: "The name of the Index.",
				Type:                types.StringType,
				Required:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"data_type": {
				MarkdownDescription: "The type of data the index holds. Possible values: `event` or `metric`.",
				Type:                types.StringType,
				Required:            true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf(validator.AllowedIndexTypes()...),
				},
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"searchable_days": {
				MarkdownDescription: "Number of days the index is searchable..",
				Type:                types.Int64Type,
				Computed:            true,
				Optional:            true,
			},
			"max_data_size_mb": {
				MarkdownDescription: "The maximum size of the index in megabytes.",
				Type:                types.Int64Type,
				Computed:            true,
				Optional:            true,
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

func (r *IndexResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IndexResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *Index

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	indexDefinition := splunkacs.IndexCreateRequest{
		Name:           data.Name.ValueString(),
		DataType:       data.DataType.ValueString(),
		SearchableDays: int(data.SearchableDays.ValueInt64()),
		MaxDataSizeMb:  int(data.MaxDataSizeMb.ValueInt64()),
	}

	tflog.Warn(ctx, "about to attempt creating an Index resource")
	indexResp, _, err := r.client.CreateIndex(indexDefinition)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while creating Index", err.Error())
		return
	}

	indexWaitResp, err := waitIndexPropagation(ctx, r.client, indexResp.Name, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while waiting for Index", err.Error())
		return
	}

	data.Name = types.StringValue(indexWaitResp.Name)
	data.DataType = types.StringValue(indexWaitResp.DataType)
	data.SearchableDays = types.Int64Value(int64(indexWaitResp.SearchableDays))
	data.MaxDataSizeMb = types.Int64Value(int64(indexWaitResp.MaxDataSizeMb))
	data.TotalEventCount = types.StringValue(indexWaitResp.TotalEventCount)
	data.TotalRawSizeMb = types.StringValue(indexWaitResp.TotalRawSizeMb)

	data.Id = types.StringValue(indexWaitResp.Name)

	tflog.Trace(ctx, "created an Index resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *Index

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	indexResp, _, err := r.client.GetIndex(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Index", err.Error())
		return
	}

	data.Name = types.StringValue(indexResp.Name)
	data.DataType = types.StringValue(indexResp.DataType)
	data.SearchableDays = types.Int64Value(int64(indexResp.SearchableDays))
	data.MaxDataSizeMb = types.Int64Value(int64(indexResp.MaxDataSizeMb))
	data.TotalEventCount = types.StringValue(indexResp.TotalEventCount)
	data.TotalRawSizeMb = types.StringValue(indexResp.TotalRawSizeMb)

	data.Id = types.StringValue(indexResp.Name)

	tflog.Trace(ctx, "read an Index resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *Index

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	indexUpdateRequest := splunkacs.IndexUpdateRequest{
		SearchableDays: int(data.SearchableDays.ValueInt64()),
		MaxDataSizeMb:  int(data.MaxDataSizeMb.ValueInt64()),
	}

	tflog.Info(ctx, "About to send update")
	tflog.Info(ctx, fmt.Sprintf("%v\n", indexUpdateRequest))

	indexUpdateResp, _, err := r.client.UpdateIndex(data.Name.ValueString(), indexUpdateRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while updating Index", err.Error())
		return
	}

	expectedState := splunkacs.Index{
		Name:           data.Name.ValueString(),
		DataType:       data.DataType.ValueString(),
		SearchableDays: int(data.SearchableDays.ValueInt64()),
		MaxDataSizeMb:  int(data.MaxDataSizeMb.ValueInt64()),
	}

	indexWaitResp, err := waitIndexPropagation(ctx, r.client, indexUpdateResp.Name, &expectedState)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while waiting for Index", err.Error())
		return
	}

	data.Name = types.StringValue(indexWaitResp.Name)
	data.DataType = types.StringValue(indexWaitResp.DataType)
	data.SearchableDays = types.Int64Value(int64(indexWaitResp.SearchableDays))
	data.MaxDataSizeMb = types.Int64Value(int64(indexWaitResp.MaxDataSizeMb))
	data.TotalEventCount = types.StringValue(indexWaitResp.TotalEventCount)
	data.TotalRawSizeMb = types.StringValue(indexWaitResp.TotalRawSizeMb)

	data.Id = types.StringValue(indexWaitResp.Name)

	tflog.Trace(ctx, "updated an Index resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IndexResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *Index

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, _, err := r.client.DeleteIndex(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unexpected error while deleting Index", err.Error())
		return
	}
}

func (r *IndexResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func waitIndexPropagation(ctx context.Context, client *splunkacs.SplunkAcsClient, indexName string, expectedState *splunkacs.Index) (*splunkacs.IndexGetResponse, error) {
	// TODO: Get rid of the for loop. Technically the timeouts should cover for us and we can fo a while true
	// TODO: Add logging inside for each iteration in the loop
	// TODO: How do I do this using the native framework? Seems to be possible in SDKv2...

	i := 0
	retries := 20
	for i < retries {
		tflog.Info(ctx, fmt.Sprintf("waiting for Index to become eventually consistent. Retry: %d\n", i))
		indexResp, httpResp, err := client.GetIndex(indexName)
		if err != nil && httpResp.StatusCode != 404 {
			tflog.Error(ctx, "encountered an unexpected error while waiting for Index to become eventually consistent")
			return nil, err
		} else if err != nil && httpResp.StatusCode == 404 {
			i++
			time.Sleep(10 * time.Second)
			continue
		}
		// We got a valid response from the API, now if expectedState was passed, time to compare if the actual and expected states are identical
		if expectedState != nil {
			actualState := splunkacs.Index{
				Name:           indexResp.Name,
				DataType:       indexResp.DataType,
				SearchableDays: indexResp.SearchableDays,
				MaxDataSizeMb:  indexResp.MaxDataSizeMb,
			}
			result := *expectedState == actualState
			tflog.Info(ctx, fmt.Sprintf("found valid response and expected state - comparing results. Result: %v\n", result))
			tflog.Info(ctx, fmt.Sprintf("value1: %v\n", *expectedState))
			tflog.Info(ctx, fmt.Sprintf("value2: %v\n", actualState))
			if !result {
				i++
				time.Sleep(10 * time.Second)
				continue
			}
			tflog.Info(ctx, "expected and actual state match")
			return indexResp, nil
		}
		return indexResp, nil
	}
	return nil, fmt.Errorf("failed to fetch a valid Index after %d retries", retries)
}
