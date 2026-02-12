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
	cmd := fmt.Sprintf("setblock %d %d %d %s replace", x, y, z, material)
	_, err := c.client.SendCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// Deletes a block.
func (c Client) DeleteBlock(ctx context.Context, x, y, z int) error {
	cmd := fmt.Sprintf("setblock %d %d %d minecraft:air replace", x, y, z)
	_, err := c.client.SendCommand(cmd)
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
	cmd := fmt.Sprintf("summon %s %s {CustomName:'{\"text\":\"%s\"}'}", entity, position, id)
	_, err := c.client.SendCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// CreateZombie summons a zombie with common zombie-specific NBT attributes.
func (c Client) CreateZombie(
	ctx context.Context,
	position string,
	id string,
	isBaby bool,
	canBreakDoors bool,
	canPickUpLoot bool,
	persistenceRequired bool,
	health float32,
) error {
	// Helper to convert Go bool → NBT byte (0b / 1b)
	boolToByte := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}

	isBabyVal := boolToByte(isBaby)
	canBreakDoorsVal := boolToByte(canBreakDoors)
	canPickUpLootVal := boolToByte(canPickUpLoot)
	persistenceRequiredVal := boolToByte(persistenceRequired)

	// Build summon command for zombie
	// Common zombie NBT tags:
	// - IsBaby (byte): 1b if baby, 0b if adult
	// - CanBreakDoors (byte): 1b to allow breaking doors
	// - CanPickUpLoot (byte): 1b to allow picking up items
	// - PersistenceRequired (byte): 1b to prevent despawn
	// - Health (float): current health (default full health is 20.0f)
	command := fmt.Sprintf(
		`summon zombie %s {CustomName:'{"text":"%s"}',IsBaby:%db,CanBreakDoors:%db,CanPickUpLoot:%db,PersistenceRequired:%db,Health:%ff}`,
		position,
		id,
		isBabyVal,
		canBreakDoorsVal,
		canPickUpLootVal,
		persistenceRequiredVal,
		health,
	)

	_, err := c.client.SendCommand(command)
	if err != nil {
		return err
	}

	return nil
}

// Create Sheep
func (c Client) CreateSheep(ctx context.Context, position string, id string, color string, sheared bool) error {
	// Map sheep colors to their NBT integer values
	colorMap := map[string]int{
		"white":      0,
		"orange":     1,
		"magenta":    2,
		"light_blue": 3,
		"yellow":     4,
		"lime":       5,
		"pink":       6,
		"gray":       7,
		"light_gray": 8,
		"cyan":       9,
		"purple":     10,
		"blue":       11,
		"brown":      12,
		"green":      13,
		"red":        14,
		"black":      15,
	}

	// Default to white if invalid color
	colorVal, ok := colorMap[color]
	if !ok {
		colorVal = 0
	}

	// Convert sheared bool → NBT-friendly value
	shearedVal := 0
	if sheared {
		shearedVal = 1
	}

	// Build summon cmd
	cmd := fmt.Sprintf(
		`summon sheep %s {CustomName:'{"text":"%s"}',Color:%d,Sheared:%db}`,
		position, id, colorVal, shearedVal,
	)

	_, err := c.client.SendCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// Deletes an entity.
func (c Client) DeleteEntity(ctx context.Context, entity string, position string, id string) error {
	// Remove the entity.
	cmd := fmt.Sprintf("kill @e[type=%s,nbt={CustomName:'{\"text\":\"%s\"}'}]", entity, id)
	_, err := c.client.SendCommand(cmd)
	if err != nil {
		return err
	}

	// Remove the entity from inventories.
	cmd = fmt.Sprintf("clear @a %s{display:{Name:'{\"text\":\"%s\"}'}}", entity, id)
	_, err = c.client.SendCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

// GameMode names keyed by the numeric values returned by Minecraft.
var gameModeNames = map[int]string{
	0: "survival",
	1: "creative",
	2: "adventure",
	3: "spectator",
}

// /data get storage minecraft:server worldDefaultGameMode
// GetDefaultGameMode queries the server for the world’s default game mode
// and returns it as a lowercase string (e.g. "creative").
func (c Client) GetDefaultGameMode(ctx context.Context) (string, error) {
	out, err := c.client.SendCommand(`/data get storage minecraft:server worldDefaultGameMode`)
	if err != nil {
		return "", fmt.Errorf("send cmd: %w", err)
	}
	// Typical output:
	// Storage minecraft:server has the following data: {worldDefaultGameMode:1}

	// Find the last colon and take everything after it.
	parts := strings.Split(out, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected response: %q", out)
	}
	numStr := strings.TrimRight(strings.TrimSpace(parts[len(parts)-1]), "}")
	id, err := strconv.Atoi(numStr)
	if err != nil {
		return "", fmt.Errorf("parse int: %w", err)
	}
	name, ok := gameModeNames[id]
	if !ok {
		return "", fmt.Errorf("unknown game mode id %d", id)
	}
	return name, nil
}

// GetUserGameMode runs `/data get entity <name> playerGameType`
// and returns the player's current game mode as a lowercase string
// ("survival", "creative", "adventure", or "spectator").
func (c Client) GetUserGameMode(ctx context.Context, name string) (string, error) {
	out, err := c.client.SendCommand(fmt.Sprintf(`/data get entity %s playerGameType`, name))
	if err != nil {
		return "", fmt.Errorf("send cmd: %w", err)
	}
	// Look for the final colon and grab everything after it.
	parts := strings.Split(out, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected response: %q", out)
	}
	numStr := strings.TrimSpace(parts[len(parts)-1])
	id, err := strconv.Atoi(numStr)
	if err != nil {
		return "", fmt.Errorf("parse int: %w", err)
	}
	nameStr, ok := gameModeNames[id]
	if !ok {
		return "", fmt.Errorf("unknown game mode id %d", id)
	}
	return nameStr, nil
}

// Sets the default game mode
func (c Client) SetDefaultGameMode(ctx context.Context, gamemode string) error {
	var cmd string
	cmd = fmt.Sprintf(`defaultgamemode %s`, gamemode)

	_, err := c.client.SendCommand(cmd)
	return err
}

// Sets the user game mode
func (c Client) SetUserGameMode(ctx context.Context, gamemode string, name string) error {
	var cmd string
	cmd = fmt.Sprintf(`gamemode %s %s`, gamemode, name)

	_, err := c.client.SendCommand(cmd)
	return err
}

func (c Client) EnableDayLock(ctx context.Context) error {
	// 1) Lock the time to day
	if _, err := c.client.SendCommand("daylock true"); err != nil {
		return fmt.Errorf("daylock true failed: %w", err)
	}

	// 2) Immediately set the world time to day
	if _, err := c.client.SendCommand("time set day"); err != nil {
		return fmt.Errorf("time set day failed: %w", err)
	}
	return nil
}

func (c Client) DisableDayLock(ctx context.Context) error {
	var cmd string
	cmd = "daylock true"
	_, err := c.client.SendCommand(cmd)
	return err
}

// Creates operator status for the specified user name
func (c Client) CreateOp(ctx context.Context, name string) error {
	var cmd string
	cmd = fmt.Sprintf(`op %s`, name)

	_, err := c.client.SendCommand(cmd)
	return err
}

// Removes operator status for the specified user name
func (c Client) RemoveOp(ctx context.Context, name string) error {
	var cmd string
	cmd = fmt.Sprintf(`deop %s`, name)

	_, err := c.client.SendCommand(cmd)
	return err
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
	// Players can be batched in one cmd
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

// Set a boolean gamerule, e.g. keepInventory, doDaylightCycle, mobGriefing, etc.
func (c Client) SetGameRuleBool(ctx context.Context, rule string, value bool) error {
	rule = strings.TrimSpace(rule)
	if !isBoolRule(rule) {
		return fmt.Errorf("gamerule %q is not a known boolean rule", rule)
	}
	val := "false"
	if value {
		val = "true"
	}
	_, err := c.client.SendCommand(fmt.Sprintf("gamerule %s %s", rule, val))
	return err
}

// Set an integer gamerule, e.g. randomTickSpeed, maxEntityCramming, spawnRadius, playersSleepingPercentage, maxCommandChainLength.
func (c Client) SetGameRuleInt(ctx context.Context, rule string, value int) error {
	rule = strings.TrimSpace(rule)
	if !isIntRule(rule) {
		return fmt.Errorf("gamerule %q is not a known integer rule", rule)
	}
	_, err := c.client.SendCommand(fmt.Sprintf("gamerule %s %d", rule, value))
	return err
}

// Read current value as a raw string. For bool rules, returns "true"/"false"; for int rules, returns the number.
func (c Client) GetGameRule(ctx context.Context, rule string) (string, error) {
	rule = strings.TrimSpace(rule)
	// Query form: /gamerule <rule>
	out, err := c.client.SendCommand(fmt.Sprintf("gamerule %s", rule))
	if err != nil {
		return "", err
	}
	// Server usually replies with just the value, but some servers/plugins may add text.
	// Try to extract the last token that parses for ints or matches true/false.
	line := strings.TrimSpace(out)
	fields := strings.Fields(line)
	if len(fields) == 1 {
		return fields[0], nil
	}
	// Heuristic: scan from end for a bool or int-looking token.
	for i := len(fields) - 1; i >= 0; i-- {
		f := strings.TrimSpace(fields[i])
		lf := strings.ToLower(f)
		if lf == "true" || lf == "false" {
			return lf, nil
		}
		if _, err := strconv.Atoi(f); err == nil {
			return f, nil
		}
	}
	// Fallback: return raw output
	return line, nil
}

// Reset (aka "delete") a gamerule back to its vanilla default.
// Returns an error if we don't have a known default for that rule.
func (c Client) ResetGameRuleToDefault(ctx context.Context, rule string) error {
	rule = strings.TrimSpace(rule)

	if def, ok := defaultBoolRules[rule]; ok {
		return c.SetGameRuleBool(ctx, rule, def)
	}
	if def, ok := defaultIntRules[rule]; ok {
		return c.SetGameRuleInt(ctx, rule, def)
	}
	return fmt.Errorf("no known default for gamerule %q; cannot reset", rule)
}

// ---- Known rules & defaults (Java Edition) ----

// Keep this list small and pragmatic; extend as you need.
// Boolean rules (subset of commonly used)
var boolRules = map[string]struct{}{
	"announceAdvancements":       {},
	"disableElytraMovementCheck": {},
	"disablePlayerMovementCheck": {},
	"disableRaids":               {},
	"doDaylightCycle":            {},
	"doEntityDrops":              {},
	"doFireTick":                 {},
	"doInsomnia":                 {},
	"doImmediateRespawn":         {},
	"doLimitedCrafting":          {},
	"doMobLoot":                  {},
	"doMobSpawning":              {},
	"doPatrolSpawning":           {},
	"doTileDrops":                {},
	"doTraderSpawning":           {},
	"doVinesSpread":              {},
	"doWeatherCycle":             {},
	"doWardenSpawning":           {},
	"drowningDamage":             {},
	"fallDamage":                 {},
	"fireDamage":                 {},
	"forgiveDeadPlayers":         {},
	"keepInventory":              {},
	"logAdminCommands":           {},
	"mobGriefing":                {},
	"naturalRegeneration":        {},
	"reducedDebugInfo":           {},
	"sendCommandFeedback":        {},
	"showDeathMessages":          {},
	"spectatorsGenerateChunks":   {},
	"universalAnger":             {},
}

// Integer rules (subset of commonly used)
var intRules = map[string]struct{}{
	"maxCommandChainLength":     {},
	"maxEntityCramming":         {},
	"playersSleepingPercentage": {},
	"randomTickSpeed":           {},
	"spawnRadius":               {},
}

// Vanilla defaults (Java). Extend as needed.
var defaultBoolRules = map[string]bool{
	"announceAdvancements":       true,
	"disableElytraMovementCheck": false,
	"disablePlayerMovementCheck": false,
	"disableRaids":               false,
	"doDaylightCycle":            true,
	"doEntityDrops":              true,
	"doFireTick":                 true,
	"doInsomnia":                 true,
	"doImmediateRespawn":         false,
	"doLimitedCrafting":          false,
	"doMobLoot":                  true,
	"doMobSpawning":              true,
	"doPatrolSpawning":           true,
	"doTileDrops":                true,
	"doTraderSpawning":           true,
	"doVinesSpread":              true,
	"doWeatherCycle":             true,
	"doWardenSpawning":           true,
	"drowningDamage":             true,
	"fallDamage":                 true,
	"fireDamage":                 true,
	"forgiveDeadPlayers":         true,
	"keepInventory":              false,
	"logAdminCommands":           true,
	"mobGriefing":                true,
	"naturalRegeneration":        true,
	"reducedDebugInfo":           false,
	"sendCommandFeedback":        true,
	"showDeathMessages":          true,
	"spectatorsGenerateChunks":   true,
	"universalAnger":             false,
}

var defaultIntRules = map[string]int{
	"maxCommandChainLength":     65536,
	"maxEntityCramming":         24,
	"playersSleepingPercentage": 100,
	"randomTickSpeed":           3,
	"spawnRadius":               5,
}

// ---- Internals ----

func isBoolRule(rule string) bool {
	_, ok := boolRules[rule]
	return ok
}

func isIntRule(rule string) bool {
	_, ok := intRules[rule]
	return ok
}

func (c Client) FillBlock(ctx context.Context, material string, sx, sy, sz, ex, ey, ez int) error {
	cmd := fmt.Sprintf("fill %d %d %d %d %d %d %s hollow", sx, sy, sz, ex, ey, ez, material)
	_, err := c.client.SendCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}
