// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKDataSource("aws_wafv2_ip_set")
func DataSourceIPSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPSetRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"addresses": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"ip_address_version": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"scope": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
				},
			}
		},
	}
}

func dataSourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)
	name := d.Get("name").(string)

	var foundIpSet awstypes.IPSetSummary
	input := &wafv2.ListIPSetsInput{
		Scope: awstypes.Scope(d.Get("scope").(string)),
		Limit: aws.Int32(100),
	}

	for {
		resp, err := conn.ListIPSets(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 IPSets: %s", err)
		}

		if resp == nil || resp.IPSets == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 IPSets")
		}

		for _, ipSet := range resp.IPSets {
			if aws.ToString(ipSet.Name) == name {
				foundIpSet = ipSet
				break
			}
		}

		if resp.NextMarker == nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundIpSet.Id == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 IPSet not found for name: %s", name)
	}

	resp, err := conn.GetIPSet(ctx, &wafv2.GetIPSetInput{
		Id:    foundIpSet.Id,
		Name:  foundIpSet.Name,
		Scope: awstypes.Scope(d.Get("scope").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 IPSet: %s", err)
	}

	if resp == nil || resp.IPSet == nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 IPSet")
	}

	d.SetId(aws.ToString(resp.IPSet.Id))
	d.Set("arn", resp.IPSet.ARN)
	d.Set("description", resp.IPSet.Description)
	d.Set("ip_address_version", resp.IPSet.IPAddressVersion)

	if err := d.Set("addresses", flex.FlattenStringValueList(resp.IPSet.Addresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting addresses: %s", err)
	}

	return diags
}
