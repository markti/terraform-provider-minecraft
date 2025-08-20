# Places a red bed: foot at (10,64,10), head at (11,64,10) because direction = east
resource "minecraft_bed" "b1" {
  material = "minecraft:red_bed"
  position = {
    x = -198
    y = 66
    z = -195
  }
  direction = "east" # HEAD is +1 in this direction
  # occupied = false  # optional
}

# Another example: bed facing north (head at z-1)
resource "minecraft_bed" "guest" {
  material = "minecraft:blue_bed"
  position = {
    x = -198
    y = 66
    z = -195
  }
  direction = "north"
}
