// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/oam"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccObservabilityAccessManagerSinkDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	ctx := acctest.Context(t)
	var sink oam.GetSinkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_oam_sink.test"
	resourceName := "aws_oam_sink.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ObservabilityAccessManagerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAccessManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSinkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSinkDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSinkExists(ctx, dataSourceName, &sink),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "sink_id", resourceName, "sink_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sink_identifier", resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccSinkDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource aws_oam_sink "test" {
  name = %[1]q

  tags = {
    key1 = "value1"
  }
}

data aws_oam_sink "test" {
  sink_identifier = aws_oam_sink.test.arn
}
`, rName)
}
