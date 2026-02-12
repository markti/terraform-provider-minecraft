package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestFillResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccFillResourceConfig("minecraft:stone", 0, 0, 0, 1, 1, 1),
		},
		{
			Config: testAccFillResourceConfig("minecraft:dirt", 1, 1, 1, 2, 2, 2),
		},
		{
			Config: testAccFillResourceConfig("minecraft:oak_planks", 2, 2, 2, 3, 3, 3),
		},
	})
}

func testAccFillResourceConfig(material string, sx, sy, sz, ex, ey, ez int) string {
	return testProviderConfig() + `
resource "minecraft_fill" "test" {
  material = "` + material + `"
  start = {
    x = ` + fmt.Sprintf("%d", sx) + `
    y = ` + fmt.Sprintf("%d", sy) + `
    z = ` + fmt.Sprintf("%d", sz) + `
  }
  end = {
    x = ` + fmt.Sprintf("%d", ex) + `
    y = ` + fmt.Sprintf("%d", ey) + `
    z = ` + fmt.Sprintf("%d", ez) + `
  }
}
`
}
