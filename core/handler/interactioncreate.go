package handler

import (
	"Cardinal"
	"Cardinal/core/commands"
	"Cardinal/core/tools"
	"context"
	"github.com/bwmarrin/discordgo"
	"time"
)

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member.User.Bot {
		return
	}
	ctx := context.Background()
	db := Cardinal.DatabaseConnect()
	if i.Type == discordgo.InteractionMessageComponent {
		data := i.MessageComponentData()
		switch data.CustomID {
		case "attack":
			time.Sleep(1 * time.Second)
			tools.ClearInteractionMessage(s, i)
			break
		case "block":
			time.Sleep(1 * time.Second)
			tools.ClearInteractionMessage(s, i)
			break
		case "dodge":
			time.Sleep(1 * time.Second)
			tools.ClearInteractionMessage(s, i)
			break
		case "skillselectbtn":
			time.Sleep(1 * time.Second)
			tools.ClearInteractionMessage(s, i)
			break
		}
	} else {
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
}
