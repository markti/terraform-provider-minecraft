---
description: Lock or unlock the in-game time to permanent daytime on a Minecraft Java server.
page_title: minecraft_daylock Resource - terraform-provider-minecraft
---

# minecraft_daylock (Resource)

Manages the daylock setting for a Minecraft Java server.

This resource allows you to:

- **Enable** permanent daytime by locking the time of day.
- **Disable** the lock to allow the normal day/night cycle.

## Example Usage

### Enable Daylock

```hcl
resource "minecraft_daylock" "default" {
  enabled = true
}
```

### Disable Daylock

```hcl
resource "minecraft_daylock" "default" {
  enabled = false
}
```

## Argument Reference

- **enabled** (Required, Boolean)\
  Set to `true` to lock the world at daytime, or `false` to restore the normal day/night cycle.

## Attribute Reference

- **id** (Computed, String)\
  Unique resource ID. Typically `"default"` when managing the global server setting.
