package handler

import (
	//    "Raphael/core/database"
	"Raphael/core/commands"
	"github.com/bwmarrin/discordgo"
)

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member.User.Bot {
		return
	}
	//	db := database.Connect()
	data := i.ApplicationCommandData()
	switch data.Name {
	case "ping":
		commands.Ping(s, i)
	}
}
