package minecraft

import (
	"context"
	"fmt"
)

func (c Client) BanPlayer(ctx context.Context, player string, reason string) error {
	cmd := fmt.Sprintf("ban %s", player)
	if reason != "" {
		cmd = fmt.Sprintf("ban %s %s", player, reason)
	}
	_, err := c.client.SendCommand(cmd)
	return err
}
