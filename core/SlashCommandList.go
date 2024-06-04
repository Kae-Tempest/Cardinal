package main

import (
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
	_, err := s.ApplicationCommandBulkOverwrite(os.Getenv("APP_ID"), os.Getenv("GUILD_ID"), commands)

	if err != nil {
		// Handle the error
	}
}
