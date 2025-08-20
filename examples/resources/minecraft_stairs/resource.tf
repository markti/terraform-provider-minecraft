# Straight stair facing east (defaults are facing=north, half=bottom, shape=straight, waterlogged=false)
resource "minecraft_stairs" "s1" {
  material = "minecraft:oak_stairs"
  position = {
    x = -198
    y = 66
    z = -195
  }

  facing      = "east"
  half        = "bottom"
  shape       = "straight"
  waterlogged = false
}

# Upside-down inner corner stair (south, inner_left)
resource "minecraft_stairs" "corner" {
  material = "minecraft:stone_brick_stairs"
  position = {
    x = -198
    y = 66
    z = -195
  }

  facing = "south"
  half   = "top"
  shape  = "inner_left"
}

# Waterlogged stair (for underwater builds)
resource "minecraft_stairs" "wet" {
  material = "minecraft:dark_oak_stairs"
  position = {
    x = -198
    y = 66
    z = -195
  }

  facing      = "west"
  waterlogged = true
}
