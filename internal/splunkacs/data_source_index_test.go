package splunkacs

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIndexDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "splunkacs_index" "test" {
	name = "splunkacs-index-ds-ci"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.splunkacs_index.test", "name", "splunkacs-index-ds-ci"),
					resource.TestCheckResourceAttr("data.splunkacs_index.test", "data_type", "event"),
					resource.TestCheckResourceAttr("data.splunkacs_index.test", "searchable_days", "30"),
					resource.TestCheckResourceAttr("data.splunkacs_index.test", "max_data_size_mb", "0"),
					resource.TestCheckResourceAttr("data.splunkacs_index.test", "total_event_count", "0"),
					resource.TestCheckResourceAttr("data.splunkacs_index.test", "total_raw_size_mb", "0"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.splunkacs_index.test", "id", "splunkacs-index-ds-ci"),
				),
			},
		},
	})
}
