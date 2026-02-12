package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestStairsResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccStairsResourceConfig("minecraft:oak_stairs", "north", "bottom", "straight", 8, 64, 8),
		},
		{
			Config: testAccStairsResourceConfig("minecraft:stone_brick_stairs", "east", "top", "inner_left", 9, 64, 9),
		},
		{
			Config: testAccStairsResourceConfig("minecraft:birch_stairs", "south", "bottom", "outer_right", 10, 64, 10),
		},
	})
}

func testAccStairsResourceConfig(material, facing, half, shape string, x, y, z int) string {
	return testProviderConfig() + `
resource "minecraft_stairs" "test" {
  material = "` + material + `"
  facing   = "` + facing + `"
  half     = "` + half + `"
  shape    = "` + shape + `"
` + testBlockPositionAttributes("position", x, y, z) + `
}
`
}
