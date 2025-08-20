# Villager farmer, named, that won't despawn
resource "minecraft_entity" "farmer" {
  type     = "minecraft:villager"
  position = { x = 0, y = 64, z = 5 }
}

# Zombie guard with sword + helmet
resource "minecraft_entity" "zombie_guard" {
  type     = "minecraft:zombie"
  position = { x = -2, y = 64, z = 5 }
}

# Blue sheep
resource "minecraft_entity" "sheep_blue" {
  type = "minecraft:sheep"
  position = {
    x = -198
    y = 66
    z = -195
  }
}
