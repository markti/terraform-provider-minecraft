package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestChestResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccChestResourceConfig("single", false, false, 9, 64, 9),
		},
		{
			Config: testAccChestResourceConfig("single", true, true, 10, 64, 10),
		},
		{
			Config: testAccChestResourceConfig("double", false, true, 11, 64, 11),
		},
	})
}

func testAccChestResourceConfig(size string, trapped, waterlogged bool, x, y, z int) string {
	return testProviderConfig() + `
resource "minecraft_chest" "test" {
  size        = "` + size + `"
  trapped     = ` + fmt.Sprintf("%t", trapped) + `
  waterlogged = ` + fmt.Sprintf("%t", waterlogged) + `
` + testBlockPositionAttributes("position", x, y, z) + `
}
`
}
