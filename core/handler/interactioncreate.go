package handler

import (
	"Cardinal"
	"Cardinal/core/commands"
	"context"
	"github.com/bwmarrin/discordgo"
)

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member.User.Bot {
		return
	}
	ctx := context.Background()
	db := Cardinal.DatabaseConnect()
	data := i.ApplicationCommandData()
	switch data.Name {
	case "ping":
		commands.Ping(s, i)
	case "setup":
		commands.Setup(ctx, s, i, db)
	case "move":
		commands.Move(ctx, s, i, db)
	case "harvest":
		commands.Harvest(ctx, s, i, db)
	case "hunt":
		commands.Hunt(ctx, s, i, db)
	}
}
