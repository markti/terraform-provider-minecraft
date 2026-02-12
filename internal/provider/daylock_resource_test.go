package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDaylockResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccDaylockResourceConfig(true),
		},
		{
			Config: testAccDaylockResourceConfig(false),
		},
		{
			Config: testAccDaylockResourceConfig(true),
		},
	})
}

func testAccDaylockResourceConfig(enabled bool) string {
	return testProviderConfig() + `
resource "minecraft_daylock" "test" {
  enabled = ` + fmt.Sprintf("%t", enabled) + `
}
`
}
