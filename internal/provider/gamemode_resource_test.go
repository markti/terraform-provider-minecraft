package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGamemodeResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccGamemodeResourceConfig("creative", ""),
		},
		{
			Config: testAccGamemodeResourceConfig("survival", "Steve"),
		},
		{
			Config: testAccGamemodeResourceConfig("adventure", "Alex"),
		},
	})
}

func testAccGamemodeResourceConfig(mode, player string) string {
	playerLine := ""
	if player != "" {
		playerLine = `  player = "` + player + `"`
	}
	return testProviderConfig() + `
resource "minecraft_gamemode" "test" {
  mode = "` + mode + `"
` + playerLine + `
}
`
}
