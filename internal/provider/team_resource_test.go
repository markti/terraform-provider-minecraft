package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTeamResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccTeamResourceConfig("blue"),
		},
		{
			Config: testAccTeamResourceConfig("red"),
		},
		{
			Config: testAccTeamResourceConfig("green"),
		},
	})
}

func testAccTeamResourceConfig(name string) string {
	return testProviderConfig() + `
resource "minecraft_team" "test" {
  name = "` + name + `"
}
`
}
