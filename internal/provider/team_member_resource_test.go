package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTeamMemberResourcePlan(t *testing.T) {
	testAccOrUnitTest(t, []resource.TestStep{
		{
			Config: testAccTeamMemberResourceConfig("blue", "Steve"),
		},
		{
			Config: testAccTeamMemberResourceConfig("red", "Alex"),
		},
		{
			Config: testAccTeamMemberResourceConfig("green", "Notch"),
		},
	})
}

func testAccTeamMemberResourceConfig(teamName, player string) string {
	return testProviderConfig() + `
resource "minecraft_team" "test" {
  name = "` + teamName + `"
}

resource "minecraft_team_member" "test" {
  team   = minecraft_team.test.name
  player = "` + player + `"
}
`
}
