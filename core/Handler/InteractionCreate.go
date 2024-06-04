package Handler

import "github.com/bwmarrin/discordgo"

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member.User.Bot {
		return
	}

	data := i.ApplicationCommandData()
	switch data.Name {
	case "ping":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong! 🏓",
			},
		})

		if err != nil {
			s.ChannelMessageSend(i.ChannelID, "An error occurred while executing this command.")
		}
	}

}
