package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBlockResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccBlockResourceConfig("minecraft:stone", 1, 2, 3),
		},
		{
			Config: testAccBlockResourceConfig("minecraft:dirt", 2, 3, 4),
		},
		{
			Config: testAccBlockResourceConfig("minecraft:oak_planks", 3, 4, 5),
		},
	})
}

func testAccBlockResourceConfig(material string, x, y, z int) string {
	return testProviderConfig() + `
resource "minecraft_block" "test" {
  material = "` + material + `"
` + testBlockPositionAttributes("position", x, y, z) + `
}
`
}
