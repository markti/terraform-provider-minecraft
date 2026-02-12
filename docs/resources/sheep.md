---
description: Summon and manage a sheep entity in a Minecraft Java
  server.
page_title: minecraft_sheep Resource - terraform-provider-minecraft
---

# minecraft_sheep (Resource)

Manages a sheep entity in a Minecraft Java server.

This resource allows you to:

- **Summon** a sheep at a given set of coordinates.
- **Customize** its wool color and initial sheared state.
- **Destroy** the sheep when the resource is removed.

## Example Usage

### Pink Sheep

``` hcl
resource "minecraft_sheep" "pink" {
  position = {
    x = -198
    y = 66
    z = -195
  }
  color   = "pink"
  sheared = false
}
```

### Green Sheared Sheep

``` hcl
resource "minecraft_sheep" "shorny" {
  position = {
    x = 100
    y = 65
    z = 200
  }
  color   = "green"
  sheared = true
}
```

## Argument Reference

- **position** (Required, Block)\
    The coordinates where the sheep will be summoned. All fields are
    required:

  - **x** (Number) -- X coordinate.
  - **y** (Number) -- Y coordinate.
  - **z** (Number) -- Z coordinate.

- **color** (Required, String)\
    The wool color of the sheep. One of:\
    `white, orange, magenta, light_blue, yellow, lime, pink, gray, light_gray, cyan, purple, blue, brown, green, red, black`.

- **sheared** (Optional, Boolean)\
    Whether the sheep is summoned in a sheared state. Defaults to
    `false`.

## Attribute Reference

- **id** (Computed, String)\
    A stable UUID used to tag and identify the sheep in the Minecraft
    world.
