package handler

import (
	"Raphael/core/commands"
	"Raphael/core/database"
	"github.com/bwmarrin/discordgo"
)

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member.User.Bot {
		return
	}
	db := database.Connect()
	data := i.ApplicationCommandData()
	switch data.Name {
	case "ping":
		commands.Ping(s, i)
	case "setup":
		commands.Setup(s, i, db)

	}
}
