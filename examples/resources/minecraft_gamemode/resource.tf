# creative, survival, adventure, spectator


# set the default gamemode
resource "minecraft_gamemode" "default" {
  mode   = "survival"
}

# sets the game mode for specific user
resource "minecraft_gamemode" "markti" {
  mode   = "creative"
  player = "markti22"
}
