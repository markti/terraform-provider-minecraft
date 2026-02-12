package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOpResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccOpResourceConfig("Steve"),
		},
		{
			Config: testAccOpResourceConfig("Alex"),
		},
		{
			Config: testAccOpResourceConfig("Notch"),
		},
	})
}

func testAccOpResourceConfig(player string) string {
	return testProviderConfig() + `
resource "minecraft_op" "test" {
  player = "` + player + `"
}
`
}
