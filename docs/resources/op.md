---
description: Grant or revoke operator (op) status for a Minecraft Java server player.
page_title: minecraft_op Resource - terraform-provider-minecraft
---

# minecraft_op (Resource)

Manages operator privileges for a player on a Minecraft Java server.

This resource allows you to:

- **Promote** a player to server operator (op), granting elevated
  administrative permissions.
- **Demote** a player by removing operator status when the resource is
  destroyed.

## Example Usage

### Grant Operator Status

```hcl
resource "minecraft_op" "markti" {
  player = "markti"
}
```

### Revoke Operator Status

Remove the resource from your configuration or run `terraform destroy`
to revoke operator status.

## Argument Reference

- **player** (Required, String)\
  The exact username of the player to grant operator privileges.

## Attribute Reference

- **id** (Computed, String)\
  Unique resource ID in the format `player:<name>`.
