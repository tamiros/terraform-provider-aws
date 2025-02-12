// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_contributor_insight_rule", name="Contributor Insight Rule")
// @Tags(identifierAttribute="rule_name")
// @Testing(importStateIdFunc="testAccContributorInsightRuleImportStateIDFunc")
// @Testing(importStateIdAttribute="rule_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatch;cloudwatch.DescribeInsightRulesOutput")
func newResourceContributorInsightRule(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceContributorInsightRule{}

	return r, nil
}

const (
	ResNameContributorInsightRule = "Contributor Insight Rule"
)

type resourceContributorInsightRule struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
}

func (r *resourceContributorInsightRule) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_cloudwatch_contributor_insight_rule"
}

func (r *resourceContributorInsightRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"rule_definition": schema.StringAttribute{
				Required: true,
			},
			"rule_name": schema.StringAttribute{
				Required: true,
			},
			"rule_state": schema.StringAttribute{
				Required: true,
			},
			// "schema": schema.StringAttribute{
			// 	Computed: true,

			// },
			// "managed_rule": schema.BoolAttribute{
			// 	Computed: true,
			// },
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *resourceContributorInsightRule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var plan resourceContributorInsightRuleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &cloudwatch.PutInsightRuleInput{
		RuleDefinition: plan.RuleDefinition.ValueStringPointer(),
		RuleName:       plan.RuleName.ValueStringPointer(),
		RuleState:      plan.RuleState.ValueStringPointer(),
	}

	in.Tags = getTagsIn(ctx)

	out, err := conn.PutInsightRule(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionCreating, ResNameContributorInsightRule, plan.RuleName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionCreating, ResNameContributorInsightRule, plan.RuleName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceContributorInsightRule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state resourceContributorInsightRuleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findContributorInsightRuleByName(ctx, conn, state.RuleName.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionSetting, ResNameContributorInsightRule, state.RuleName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceContributorInsightRule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state resourceContributorInsightRuleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &cloudwatch.DeleteInsightRulesInput{
		RuleNames: []string{state.RuleName.ValueString()},
	}

	_, err := conn.DeleteInsightRules(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionDeleting, ResNameContributorInsightRule, state.RuleName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceContributorInsightRule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("rule_name"), req, resp)
}

func (r *resourceContributorInsightRule) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findContributorInsightRuleByName(ctx context.Context, conn *cloudwatch.Client, name string) (*awstypes.InsightRule, error) {
	input := &cloudwatch.DescribeInsightRulesInput{}
	out, err := findContributorInsightRules(ctx, conn, input, name)
	if err != nil {
		return nil, err
	}

	// if out == nil {
	// 	return nil, tfresource.NewEmptyResultError(input)
	// }

	return tfresource.AssertSingleValueResult(out)
}

func findContributorInsightRules(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.DescribeInsightRulesInput, name string) ([]awstypes.InsightRule, error) {
	var output []awstypes.InsightRule

	log.Printf("[WARN]input: %+v", input)

	paginator := cloudwatch.NewDescribeInsightRulesPaginator(conn, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.InsightRules {
			if aws.ToString(v.Name) == name {
				output = append(output, v)
			}
		}

		output = append(output, page.InsightRules...)
	}

	return output, nil
}

type resourceContributorInsightRuleData struct {
	RuleDefinition types.String `tfsdk:"rule_definition"`
	RuleName       types.String `tfsdk:"rule_name"`
	RuleState      types.String `tfsdk:"rule_state"`
	Tags           tftags.Map   `tfsdk:"tags"`
	TagsAll        tftags.Map   `tfsdk:"tags_all"`
}
