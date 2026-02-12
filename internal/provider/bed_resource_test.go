package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBedResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccBedResourceConfig("minecraft:red_bed", "north", 7, 64, 7),
		},
		{
			Config: testAccBedResourceConfig("minecraft:blue_bed", "east", 8, 64, 8),
		},
		{
			Config: testAccBedResourceConfig("minecraft:white_bed", "south", 9, 64, 9),
		},
	})
}

func testAccBedResourceConfig(material, direction string, x, y, z int) string {
	return testProviderConfig() + `
resource "minecraft_bed" "test" {
  material  = "` + material + `"
  direction = "` + direction + `"
` + testBlockPositionAttributes("position", x, y, z) + `
}
`
}
