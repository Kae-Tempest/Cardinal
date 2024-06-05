package main

import (
	"log/slog"

	//    "log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
)

func loadInteractionCommand(s *discordgo.Session) {

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Replies with Pong!",
		},
	}

	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, os.Getenv("GUILD_ID"), commands)

	if err != nil {
		slog.Error("Error creating interaction command: ", err)
	}
}
