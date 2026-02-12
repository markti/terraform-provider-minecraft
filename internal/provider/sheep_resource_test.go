package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSheepResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccSheepResourceConfig("white", 10, 64, 10),
		},
		{
			Config: testAccSheepResourceConfig("black", 11, 64, 11),
		},
		{
			Config: testAccSheepResourceConfig("red", 12, 64, 12),
		},
	})
}

func testAccSheepResourceConfig(color string, x, y, z int) string {
	return testProviderConfig() + `
resource "minecraft_sheep" "test" {
  color = "` + color + `"
` + testBlockPositionAttributes("position", x, y, z) + `
}
`
}
