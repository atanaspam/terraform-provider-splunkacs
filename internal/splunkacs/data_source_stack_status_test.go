package splunkacs

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStackStatusDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + `
data "splunkacs_stack_status" "test" {
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.splunkacs_stack_status.test", "type", "victoria"),
					resource.TestCheckResourceAttr("data.splunkacs_stack_status.test", "version", "9.0.2208.4"), // this will probably eventually fail. Not sure what is the best way to test this, perhaps regex?

					// Verify placeholder id attribute
					// resource.TestCheckResourceAttr("data.splunkacs_stack_status.test", "id", "??"), // This will leak my test instance id which would not be a good idea ;) For now we will have to trust the code
				),
			},
		},
	})
}
