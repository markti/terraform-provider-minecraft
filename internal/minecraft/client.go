package minecraft

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/seeruk/minecraft-rcon/rcon"
)

type Client struct {
	client *rcon.Client
}

type Player struct {
}

func New(address string, password string) (*Client, error) {
	addressParts := strings.Split(address, ":")
	host := addressParts[0]
	port, err := strconv.Atoi(addressParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port %s", addressParts[1])
	}

	client, err := rcon.NewClient(host, port, password)
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

// Get a player.
func (c Client) GetPlayer(ctx context.Context, name string) error {
	return nil
}

// Creates a block.
func (c Client) CreateBlock(ctx context.Context, material string, x, y, z int) error {
	command := fmt.Sprintf("setblock %d %d %d %s replace", x, y, z, material)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}

// Deletes a block.
func (c Client) DeleteBlock(ctx context.Context, x, y, z int) error {
	command := fmt.Sprintf("setblock %d %d %d minecraft:air replace", x, y, z)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}

// CreateStairs places a stairs block (e.g., "minecraft:oak_stairs") with orientation.
func (c Client) CreateStairs(ctx context.Context, material string, x, y, z int, facing, half, shape string, waterlogged bool) error {
	cmd := fmt.Sprintf(
		`setblock %d %d %d %s[facing=%s,half=%s,shape=%s,waterlogged=%t] replace`,
		x, y, z, material, facing, half, shape, waterlogged,
	)
	_, err := c.client.SendCommand(cmd)
	return err
}

// Creates an entity.
func (c Client) CreateEntity(ctx context.Context, entity string, position string, id string) error {
	command := fmt.Sprintf("summon %s %s {CustomName:'{\"text\":\"%s\"}'}", entity, position, id)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}

// Deletes an entity.
func (c Client) DeleteEntity(ctx context.Context, entity string, position string, id string) error {
	// Remove the entity.
	command := fmt.Sprintf("kill @e[type=%s,nbt={CustomName:'{\"text\":\"%s\"}'}]", entity, id)
	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	// Remove the entity from inventories.
	command = fmt.Sprintf("clear @a %s{display:{Name:'{\"text\":\"%s\"}'}}", entity, id)
	_, err = c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}

// Creates a team with a given name and optional display name.
func (c Client) CreateTeam(ctx context.Context, name string, displayName string) error {
	var cmd string
	if displayName != "" {
		cmd = fmt.Sprintf(`team add %s "%s"`, name, displayName)
	} else {
		cmd = fmt.Sprintf(`team add %s`, name)
	}

	_, err := c.client.SendCommand(cmd)
	return err
}

// Deletes a team by name.
func (c Client) DeleteTeam(ctx context.Context, name string) error {
	cmd := fmt.Sprintf("team remove %s", name)
	_, err := c.client.SendCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// --- New: Set options via /team modify
// Color: e.g. white, gray, dark_gray, black, red, dark_red, gold, yellow, green, dark_green,
// aqua, dark_aqua, blue, dark_blue, light_purple, dark_purple
func (c Client) SetTeamColor(ctx context.Context, name, color string) error {
	color = strings.ToLower(color)
	_, err := c.client.SendCommand(fmt.Sprintf("team modify %s color %s", name, color))
	return err
}

func (c Client) SetTeamFriendlyFire(ctx context.Context, name string, enabled bool) error {
	val := "true"
	if !enabled {
		val = "false"
	}
	_, err := c.client.SendCommand(fmt.Sprintf("team modify %s friendlyFire %s", name, val))
	return err
}

func (c Client) SetTeamSeeFriendlyInvisibles(ctx context.Context, name string, enabled bool) error {
	val := "true"
	if !enabled {
		val = "false"
	}
	_, err := c.client.SendCommand(fmt.Sprintf("team modify %s seeFriendlyInvisibles %s", name, val))
	return err
}

// Nametag visibility: always | never | hideForOtherTeams | hideForOwnTeam
func (c Client) SetTeamNametagVisibility(ctx context.Context, name, mode string) error {
	mode = strings.TrimSpace(mode)
	_, err := c.client.SendCommand(fmt.Sprintf("team modify %s nametagVisibility %s", name, mode))
	return err
}

// Collision rule: always | never | pushOtherTeams | pushOwnTeam
func (c Client) SetTeamCollisionRule(ctx context.Context, name, rule string) error {
	rule = strings.TrimSpace(rule)
	_, err := c.client.SendCommand(fmt.Sprintf("team modify %s collisionRule %s", name, rule))
	return err
}

// Display name: Minecraft accepts a text component; a plain quoted string also works.
// Safest is a simple text component.
func (c Client) SetTeamDisplayName(ctx context.Context, name, display string) error {
	escaped := strings.ReplaceAll(display, `"`, `\"`)
	cmd := fmt.Sprintf(`team modify %s displayName {"text":"%s"}`, name, escaped)
	_, err := c.client.SendCommand(cmd)
	return err
}

// Join arbitrary targets to a team (players or selectors).
// Examples:
//
//	JoinTeamTargets(ctx, "blue", "Steve")
//	JoinTeamTargets(ctx, "red", "@a[team=]")
//	JoinTeamTargets(ctx, "blue", "@e[type=minecraft:zombie,limit=5]")
func (c Client) JoinTeamTargets(ctx context.Context, team string, targets ...string) error {
	if len(targets) == 0 {
		return nil
	}
	cmd := fmt.Sprintf("team join %s %s", team, strings.Join(targets, " "))
	_, err := c.client.SendCommand(cmd)
	return err
}

// Make the given targets leave whichever team they’re in.
// Examples:
//
//	LeaveTeamTargets(ctx, "Steve")
//	LeaveTeamTargets(ctx, "@e[type=minecraft:zombie,distance=..10]")
func (c Client) LeaveTeamTargets(ctx context.Context, targets ...string) error {
	if len(targets) == 0 {
		return nil
	}
	cmd := fmt.Sprintf("team leave %s", strings.Join(targets, " "))
	_, err := c.client.SendCommand(cmd)
	return err
}

// ---------- Convenience: players by name ----------

func (c Client) JoinTeamPlayers(ctx context.Context, team string, players ...string) error {
	// Players can be batched in one command
	return c.JoinTeamTargets(ctx, team, players...)
}

func (c Client) LeaveTeamPlayers(ctx context.Context, players ...string) error {
	return c.LeaveTeamTargets(ctx, players...)
}

// ---------- Convenience: entities by stable CustomName ----------
// You mentioned you embed a UUID in the entity's CustomName when creating it.
// We can build a selector that matches that name exactly.
//
// NOTE: We escape double-quotes in the name to keep the JSON valid.
func selectorByCustomName(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `\"`)
	// Matches exact name text component
	return fmt.Sprintf(`@e[nbt={CustomName:'{"text":"%s"}'}]`, escaped)
}

func (c Client) JoinTeamEntityByName(ctx context.Context, team string, customName string) error {
	sel := selectorByCustomName(customName)
	return c.JoinTeamTargets(ctx, team, sel)
}

func (c Client) LeaveTeamEntityByName(ctx context.Context, customName string) error {
	sel := selectorByCustomName(customName)
	return c.LeaveTeamTargets(ctx, sel)
}

// ---------- Convenience: bulk entities by tag (recommended) ----------
// If you also tag entities (e.g., `tag add <id>` or in your summon NBT), selectors by tag
// are very cheap and reliable. This joins/leaves all matching entities.

func (c Client) JoinTeamEntitiesByTag(ctx context.Context, team, tag string) error {
	return c.JoinTeamTargets(ctx, team, fmt.Sprintf(`@e[tag=%s]`, tag))
}

func (c Client) LeaveTeamEntitiesByTag(ctx context.Context, tag string) error {
	return c.LeaveTeamTargets(ctx, fmt.Sprintf(`@e[tag=%s]`, tag))
}
