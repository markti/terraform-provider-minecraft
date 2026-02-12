package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGameruleResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccGameruleResourceConfig("keepInventory", "true"),
		},
		{
			Config: testAccGameruleResourceConfig("doDaylightCycle", "false"),
		},
		{
			Config: testAccGameruleResourceConfig("randomTickSpeed", "3"),
		},
	})
}

func testAccGameruleResourceConfig(name, value string) string {
	return testProviderConfig() + `
resource "minecraft_gamerule" "test" {
  name  = "` + name + `"
  value = "` + value + `"
}
`
}
