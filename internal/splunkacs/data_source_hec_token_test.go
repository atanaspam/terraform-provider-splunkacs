package splunkacs

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccHecTokenDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "splunkacs_hec_token" "test" {
	name = "splunkacs-provider-ci-p"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of coffees returned
					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "allowed_indexes.#", "1"),

					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "allowed_indexes.0", "main"),
					// resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "default_host", "??"), // I have to hardcode this for my test instace which would not be a good idea ;) For now we will have to trust the code
					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "default_index", "main"),
					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "default_source", "hec"),
					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "default_sourcetype", "_json"),
					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "disabled", "false"),
					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "use_ack", "false"),
					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "name", "splunkacs-provider-ci-p"),

					// Verify placeholder id attribute
					resource.TestCheckResourceAttr("data.splunkacs_hec_token.test", "id", "splunkacs-provider-ci-p"),
				),
			},
		},
	})
}
