---
description: Set the default server gamemode or change a specific
  player's current gamemode.
page_title: minecraft_gamemode Resource - terraform-provider-minecraft
---

# minecraft_gamemode (Resource)

Manages the game mode for a Minecraft Java server.

This resource can either:

-   Set the **world's default game mode** (what new players or respawns
    inherit), or\
-   Set the **current game mode for a specific player**.

## Example Usage

### Set World Default to Creative

``` hcl
resource "minecraft_gamemode" "default" {
  mode = "creative"
}
```

### Force a Player into Spectator

``` hcl
resource "minecraft_gamemode" "mark" {
  player = "markti"
  mode   = "spectator"
}
```

## Argument Reference

-   **mode** (Required, String)\
    Target game mode. Must be one of:\
    `survival`, `creative`, `adventure`, `spectator`.

-   **player** (Optional, String)\
    If provided, applies the mode to this specific player.\
    If omitted, the provider sets the **server's default** game mode.

## Attribute Reference

-   **id** (Computed, String)\
    Unique resource ID:
    -   `"default"` when managing the server default, or\
    -   `"player:<name>"` when targeting a specific player.
-   **previous_mode** (Computed, String)\
    Best-effort snapshot of the player's or world's prior mode at the
    time of creation or last update.\
    Used to revert on resource deletion.
