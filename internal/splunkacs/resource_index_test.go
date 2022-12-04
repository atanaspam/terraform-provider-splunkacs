package splunkacs

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIndexResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "splunkacs_index" "test" {
	name             = "splunkacs-index-rs-ci"
	data_type        = "event"
	searchable_days  = 30
	max_data_size_mb = 0
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("splunkacs_index.test", "name", "splunkacs-index-rs-ci"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "data_type", "event"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "searchable_days", "30"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "max_data_size_mb", "0"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "total_event_count", "0"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "total_raw_size_mb", "0"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("splunkacs_index.test", "id", "splunkacs-index-rs-ci"),
				),
			},
			// // Wait for index creation to propagate
			// {
			// 	PreConfig:    func() { time.Sleep(300 * time.Second) },
			// 	SkipFunc:     skip,
			// 	RefreshState: true,
			// },
			// ImportState testing
			{
				ResourceName:      "splunkacs_index.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "splunkacs_index" "test" {
	name             = "splunkacs-index-rs-ci"
	data_type        = "event"
	searchable_days  = 20
	max_data_size_mb = 1024
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("splunkacs_index.test", "name", "splunkacs-index-rs-ci"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "data_type", "event"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "searchable_days", "20"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "max_data_size_mb", "1024"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "total_event_count", "0"),
					resource.TestCheckResourceAttr("splunkacs_index.test", "total_raw_size_mb", "0"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("splunkacs_index.test", "id", "splunkacs-index-rs-ci"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
