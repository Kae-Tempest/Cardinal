package tools

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func ClearInteractionMessage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionResponseDelete(i.Interaction)
	if err != nil {
		slog.Error("Error during deleting Interaction", err)
	}
}
