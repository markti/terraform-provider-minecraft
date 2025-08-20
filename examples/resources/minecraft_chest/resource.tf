
resource "minecraft_chest" "example" {
  size        = "double" # or "single"
  trapped     = true     # optional
  waterlogged = false    # optional

  position = {
    x = 10
    y = 64
    z = 10
  }
}
