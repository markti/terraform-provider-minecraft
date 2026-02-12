package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestZombieResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccZombieResourceConfig(11, 64, 11),
		},
		{
			Config: testAccZombieResourceConfig(12, 64, 12),
		},
		{
			Config: testAccZombieResourceConfig(13, 64, 13),
		},
	})
}

func testAccZombieResourceConfig(x, y, z int) string {
	return testProviderConfig() + `
resource "minecraft_zombie" "test" {
  position = {
    x = ` + fmt.Sprintf("%d", x) + `
    y = ` + fmt.Sprintf("%d", y) + `
    z = ` + fmt.Sprintf("%d", z) + `
  }
}
`
}
