
resource "minecraft_zombie" "z1" {
  position = {
    x = 595
    y = 64
    z = 198
  }
  is_baby              = false
  can_break_doors      = false
  can_pick_up_loot     = true
  persistence_required = true
  health               = 20.0
}
