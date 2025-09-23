resource "minecraft_summon_villager" "farmer" {
  x = 100
  y = 64
  z = 200
  
  data_tag = jsonencode({
    VillagerData = {
      profession = "farmer"
      level      = 2
      type       = "plains"
    }
  })
}

resource "minecraft_summon_villager" "librarian" {
  x = 101
  y = 64
  z = 201
  
  data_tag = jsonencode({
    VillagerData = {
      profession = "librarian"
      level      = 5
      type       = "desert"
    }
    CustomNameVisible = true
  })
}

resource "minecraft_summon_villager" "legacy_trader" {
  x = 102
  y = 64
  z = 202
  
  # Legacy format using numeric profession IDs
  data_tag = jsonencode({
    Profession   = 1  # Librarian
    Career       = 1  # Librarian
    CareerLevel  = 3
  })
}

resource "minecraft_summon_villager" "basic_villager" {
  x = 103
  y = 64
  z = 203
  
  # Basic villager without additional data tags
}