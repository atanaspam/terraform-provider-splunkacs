package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccHecTokenResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "splunkacs_hec_token" "test" {
	name               = "splunkacs-provider-ci"
	allowed_indexes    = ["main"]
	default_index      = "main"
	default_source     = "hec"
	default_sourcetype = "_json"
	disabled           = false
	use_ack            = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of items
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "allowed_indexes.#", "1"),

					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "allowed_indexes.0", "main"),
					// resource.TestCheckResourceAttr("splunkacs_hec_token.test", "default_host", "??"), // I have to hardcode this for my test instace which would not be a good idea ;) For now we will have to trust the code
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "default_index", "main"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "default_source", "hec"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "default_sourcetype", "_json"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "disabled", "false"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "use_ack", "false"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "name", "splunkacs-provider-ci"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "id", "splunkacs-provider-ci"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "splunkacs_hec_token.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "splunkacs_hec_token" "test" {
	name               = "splunkacs-provider-ci"
	allowed_indexes    = ["main"]
	default_index      = "main"
	default_source     = "hec_token"
	default_sourcetype = "_json"
	disabled           = false
	use_ack            = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "allowed_indexes.#", "1"),

					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "allowed_indexes.0", "main"),
					// resource.TestCheckResourceAttr("splunkacs_hec_token.test", "default_host", "??"), // I have to hardcode this for my test instace which would not be a good idea ;) For now we will have to trust the code
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "default_index", "main"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "default_source", "hec_token"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "default_sourcetype", "_json"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "disabled", "false"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "use_ack", "true"),
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "name", "splunkacs-provider-ci"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("splunkacs_hec_token.test", "id", "splunkacs-provider-ci"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
