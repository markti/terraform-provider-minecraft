terraform {
  required_providers {
    minecraft = {
      source  = "HashiCraft/minecraft"
      version = "0.1.0"
    }
  }
}

// Configure the provider with the RCON details of the Minecraft server
provider "minecraft" {
  address  = "localhost:27015"
  password = "password"
}

// Create a Minecraft block
resource "minecraft_block" "stone" {
  material = "minecraft:stone"

  position = {
    x = -198
    y = 66
    z = -195
  }
}

// Module that creates a cube out of Minecraft blocks
module "cube" {
  source = "./cube"

  material = "cobblestone"

  position = {
    x = -198
    y = 68
    z = -195
  }

  dimensions = {
    width  = 3
    length = 3
    height = 3
  }
}

// ban a user
resource "minecraft_ban" "example" {
  player = "Goblin"
  reason = "Griefing"
}