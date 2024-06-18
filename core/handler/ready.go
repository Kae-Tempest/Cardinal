package handler

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"strings"
)

func Ready(s *discordgo.Session, r *discordgo.Ready) {
	var categoryID = "nil"
	var textID = "nil"
	for _, guild := range r.Guilds {
		channels, getChannelsErr := s.GuildChannels(guild.ID)
		if getChannelsErr != nil {
			slog.Error("Error during getting Channels", getChannelsErr)
			return
		}
		for _, channel := range channels {
			if strings.Contains(channel.Name, "Raphael Bot") && channel.Type == discordgo.ChannelTypeGuildCategory {
				categoryID = channel.ID
			}
			if strings.Contains(channel.Name, "fights") && channel.Type == discordgo.ChannelTypeGuildText {
				textID = channel.ID
			}
		}
		if categoryID == "nil" {
			categoryChannel, categoryChannelCreateErr := s.GuildChannelCreate(guild.ID, "Raphael Bot", discordgo.ChannelTypeGuildCategory)
			if categoryChannelCreateErr != nil {
				slog.Error(fmt.Sprintf("Error during Caterogy Creation on guild: %s with ID: %s", guild.Name, guild.ID), categoryChannelCreateErr)
				return
			}
			categoryID = categoryChannel.ID
		}
		if textID == "nil" && categoryID != "nil" {
			textChannel, textChannelCreateErr := s.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
				Name:                 "fights",
				Type:                 discordgo.ChannelTypeGuildText,
				PermissionOverwrites: nil,
				ParentID:             categoryID,
			})
			if textChannelCreateErr != nil {
				slog.Error(fmt.Sprintf("Error during Text Channel Creation on guild: %s with ID: %s", guild.Name, guild.ID), textChannelCreateErr)
				return
			}
			textID = textChannel.ID
		}
	}
}
