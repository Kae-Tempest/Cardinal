package main

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"os"
)

func loadInteractionCommand(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Replies with Pong!",
		},
		{
			Name:        "setup",
			Description: "Character Creation",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "Choose your Username",
					Required:    true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "race",
					Description:  "Choose your Race",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "job",
					Description:  "Choose your Job",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "move",
			Description: "Move to one place",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "location",
					Description:  "Location where do you want to go",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		{
			Name:        "hunt",
			Description: "Hunt creatures",
		},
		{
			Name:        "harvest",
			Description: "Harvest resources",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Description:  "Choose type of Resource",
					Name:         "type",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Description:  "Choose duration",
					Name:         "time",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
	}
	_, applicationCommandErr := s.ApplicationCommandBulkOverwrite(s.State.User.ID, os.Getenv("GUILD_ID"), commands)

	if applicationCommandErr != nil {
		slog.Error("Error creating interaction command: ", applicationCommandErr)
	}
}
