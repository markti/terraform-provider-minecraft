resource "minecraft_team" "red" {
  display_name = "Red Team"
  name         = "red"
  color        = "red"
}

resource "minecraft_team_member" "markti" {
  team   = minecraft_team.red.id
  player = "markti22"
}
