// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0.

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/ehealth-co-id/terraform-provider-uptimekuma/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MonitorResource{}
var _ resource.ResourceWithImportState = &MonitorResource{}

func NewMonitorResource() resource.Resource {
	return &MonitorResource{}
}

// MonitorResource defines the resource implementation.
type MonitorResource struct {
	client *client.Client
}

// MonitorResourceModel describes the resource data model.
type MonitorResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	URL            types.String `tfsdk:"url"`
	Method         types.String `tfsdk:"method"`
	Hostname       types.String `tfsdk:"hostname"`
	Port           types.Int64  `tfsdk:"port"`
	Interval       types.Int64  `tfsdk:"interval"`
	RetryInterval  types.Int64  `tfsdk:"retry_interval"`
	ResendInterval types.Int64  `tfsdk:"resend_interval"`
	MaxRetries     types.Int64  `tfsdk:"max_retries"`
	UpsideDown     types.Bool   `tfsdk:"upside_down"`
	IgnoreTLS      types.Bool   `tfsdk:"ignore_tls"`
	MaxRedirects   types.Int64  `tfsdk:"max_redirects"`
	Body           types.String `tfsdk:"body"`
	Headers        types.String `tfsdk:"headers"`
	AuthMethod     types.String `tfsdk:"auth_method"`
	BasicAuthUser  types.String `tfsdk:"basic_auth_user"`
	BasicAuthPass  types.String `tfsdk:"basic_auth_pass"`
	Keyword        types.String `tfsdk:"keyword"`
}

func (r *MonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor"
}

func (r *MonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// ADD THIS: Top-level description for the resource documentation.
		MarkdownDescription: "Manages an Uptime Kuma monitor, allowing creation, modification, and deletion of various monitor types.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Monitor identifier.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Monitor type (http, ping, port, etc.).",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Monitor name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Monitor description.",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL to monitor (required for http, keyword monitors).",
				Optional:            true,
			},
			"method": schema.StringAttribute{
				MarkdownDescription: "HTTP method (GET, POST, etc.) for http monitors.",
				Optional:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname for ping, port, etc. monitors.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port number for port monitors.",
				Optional:            true,
			},
			"interval": schema.Int64Attribute{
				MarkdownDescription: "Check interval in seconds.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
			},
			"retry_interval": schema.Int64Attribute{
				MarkdownDescription: "Retry interval in seconds.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
			},
			"resend_interval": schema.Int64Attribute{
				MarkdownDescription: "Notification resend interval in seconds.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"max_retries": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of retries.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"upside_down": schema.BoolAttribute{
				MarkdownDescription: "Invert status (treat DOWN as UP and vice versa).",
				Optional:            true,
			},
			"ignore_tls": schema.BoolAttribute{
				MarkdownDescription: "Ignore TLS/SSL errors.",
				Optional:            true,
			},
			"max_redirects": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of redirects to follow.",
				Optional:            true,
			},
			"body": schema.StringAttribute{
				MarkdownDescription: "Request body for http monitors.",
				Optional:            true,
			},
			"headers": schema.StringAttribute{
				MarkdownDescription: "Request headers for http monitors (JSON format).",
				Optional:            true,
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method (basic, ntlm, mtls).",
				Optional:            true,
			},
			"basic_auth_user": schema.StringAttribute{
				MarkdownDescription: "Basic auth username.",
				Optional:            true,
			},
			"basic_auth_pass": schema.StringAttribute{
				MarkdownDescription: "Basic auth password.",
				Optional:            true,
				Sensitive:           true,
			},
			"keyword": schema.StringAttribute{
				MarkdownDescription: "Keyword to search for in response.",
				Optional:            true,
			},
			// where the API provides defaults if not specified by the user.
			// Also added periods to all descriptions proactively for 'godot'.
		},
	}
}

func (r *MonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *MonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare the API request
	monitor := &client.Monitor{
		Type:           client.MonitorType(data.Type.ValueString()),
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
		Interval:       int(data.Interval.ValueInt64()),
		RetryInterval:  int(data.RetryInterval.ValueInt64()),
		ResendInterval: int(data.ResendInterval.ValueInt64()),
		MaxRetries:     int(data.MaxRetries.ValueInt64()),
		UpsideDown:     data.UpsideDown.ValueBool(),
		IgnoreTLS:      data.IgnoreTLS.ValueBool(),
	}

	// Set optional fields.
	if !data.URL.IsNull() {
		monitor.URL = data.URL.ValueString()
	}

	if !data.Method.IsNull() {
		monitor.Method = data.Method.ValueString()
	}

	if !data.Hostname.IsNull() {
		monitor.Hostname = data.Hostname.ValueString()
	}

	if !data.Port.IsNull() {
		monitor.Port = int(data.Port.ValueInt64())
	}

	if !data.MaxRedirects.IsNull() {
		monitor.MaxRedirects = int(data.MaxRedirects.ValueInt64())
	}

	if !data.Body.IsNull() {
		monitor.Body = data.Body.ValueString()
	}

	if !data.Headers.IsNull() {
		monitor.Headers = data.Headers.ValueString()
	}

	if !data.AuthMethod.IsNull() {
		monitor.AuthMethod = client.AuthMethod(data.AuthMethod.ValueString())
	}

	if !data.BasicAuthUser.IsNull() {
		monitor.BasicAuthUser = data.BasicAuthUser.ValueString()
	}

	if !data.BasicAuthPass.IsNull() {
		monitor.BasicAuthPass = data.BasicAuthPass.ValueString()
	}

	if !data.Keyword.IsNull() {
		monitor.Keyword = data.Keyword.ValueString()
	}

	// Create the monitor.
	tflog.Info(ctx, "Creating monitor", map[string]interface{}{
		"name": monitor.Name,
		"type": monitor.Type,
	})

	createdMonitor, err := r.client.CreateMonitor(ctx, monitor)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create monitor: %s", err))
		return
	}

	// Update Terraform state.
	data.ID = types.Int64Value(int64(createdMonitor.ID))

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorResourceModel

	// Read Terraform prior state data into the model.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	monitorID := int(data.ID.ValueInt64())

	// Read the monitor from the API.
	monitor, err := r.client.GetMonitor(ctx, monitorID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read monitor %d: %s", monitorID, err),
		)
		return
	}

	// Update the data model.
	data.ID = types.Int64Value(int64(monitor.ID))
	data.Type = types.StringValue(string(monitor.Type))
	data.Name = types.StringValue(monitor.Name)
	data.Description = types.StringValue(monitor.Description)
	data.URL = types.StringValue(monitor.URL)
	data.Method = types.StringValue(monitor.Method)
	data.Hostname = types.StringValue(monitor.Hostname)
	data.Port = types.Int64Value(int64(monitor.Port))
	data.Interval = types.Int64Value(int64(monitor.Interval))
	data.RetryInterval = types.Int64Value(int64(monitor.RetryInterval))
	data.ResendInterval = types.Int64Value(int64(monitor.ResendInterval))
	data.MaxRetries = types.Int64Value(int64(monitor.MaxRetries))
	data.UpsideDown = types.BoolValue(monitor.UpsideDown)
	data.IgnoreTLS = types.BoolValue(monitor.IgnoreTLS)
	data.MaxRedirects = types.Int64Value(int64(monitor.MaxRedirects))
	data.Body = types.StringValue(monitor.Body)
	data.Headers = types.StringValue(monitor.Headers)
	data.AuthMethod = types.StringValue(string(monitor.AuthMethod))
	data.BasicAuthUser = types.StringValue(monitor.BasicAuthUser)
	data.BasicAuthPass = types.StringValue(monitor.BasicAuthPass)
	data.Keyword = types.StringValue(monitor.Keyword)

	// Save updated data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	monitorID := int(data.ID.ValueInt64())

	// Prepare the API requester.
	monitor := &client.Monitor{
		Type:           client.MonitorType(data.Type.ValueString()),
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
		Interval:       int(data.Interval.ValueInt64()),
		RetryInterval:  int(data.RetryInterval.ValueInt64()),
		ResendInterval: int(data.ResendInterval.ValueInt64()),
		MaxRetries:     int(data.MaxRetries.ValueInt64()),
		UpsideDown:     data.UpsideDown.ValueBool(),
		IgnoreTLS:      data.IgnoreTLS.ValueBool(),
	}

	// Set optional field.
	if !data.URL.IsNull() {
		monitor.URL = data.URL.ValueString()
	}

	if !data.Method.IsNull() {
		monitor.Method = data.Method.ValueString()
	}

	if !data.Hostname.IsNull() {
		monitor.Hostname = data.Hostname.ValueString()
	}

	if !data.Port.IsNull() {
		monitor.Port = int(data.Port.ValueInt64())
	}

	if !data.MaxRedirects.IsNull() {
		monitor.MaxRedirects = int(data.MaxRedirects.ValueInt64())
	}

	if !data.Body.IsNull() {
		monitor.Body = data.Body.ValueString()
	}

	if !data.Headers.IsNull() {
		monitor.Headers = data.Headers.ValueString()
	}

	if !data.AuthMethod.IsNull() {
		monitor.AuthMethod = client.AuthMethod(data.AuthMethod.ValueString())
	}

	if !data.BasicAuthUser.IsNull() {
		monitor.BasicAuthUser = data.BasicAuthUser.ValueString()
	}

	if !data.BasicAuthPass.IsNull() {
		monitor.BasicAuthPass = data.BasicAuthPass.ValueString()
	}

	if !data.Keyword.IsNull() {
		monitor.Keyword = data.Keyword.ValueString()
	}

	// Update the monitor.
	tflog.Info(ctx, "Updating monitor", map[string]interface{}{
		"id":   monitorID,
		"name": monitor.Name,
	})

	_, err := r.client.UpdateMonitor(ctx, monitorID, monitor)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update monitor %d: %s", monitorID, err))
		return
	}

	// Save updated data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorResourceModel

	// Read Terraform prior state data into the model.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	monitorID := int(data.ID.ValueInt64())

	// Delete the monitor.
	tflog.Info(ctx, "Deleting monitor", map[string]interface{}{
		"id": monitorID,
	})

	err := r.client.DeleteMonitor(ctx, monitorID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete monitor %d: %s", monitorID, err))
		return
	}
}

func (r *MonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Convert import ID (string) to int.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Monitor ID",
			fmt.Sprintf("Monitor ID must be a number, got: %s", req.ID),
		)
		return
	}

	// Set the ID in the state.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
