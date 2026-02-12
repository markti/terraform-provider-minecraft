package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestEntityResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccEntityResourceConfig("minecraft:armor_stand", 4, 5, 6),
		},
		{
			Config: testAccEntityResourceConfig("minecraft:text_display", 5, 6, 7),
		},
		{
			Config: testAccEntityResourceConfig("minecraft:item_frame", 6, 7, 8),
		},
	})
}

func testAccEntityResourceConfig(entityType string, x, y, z int) string {
	return testProviderConfig() + `
resource "minecraft_entity" "test" {
  type = "` + entityType + `"
` + testBlockPositionAttributes("position", x, y, z) + `
}
`
}
