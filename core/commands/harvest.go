package commands

import (
	_struct "Raphael/core/struct"
	"Raphael/core/utils"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

func Harvest(s *discordgo.Session, i *discordgo.InteractionCreate, db *pgxpool.Pool) {
	ctx := context.Background()

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		fmt.Println(data)
		var player _struct.Player
		selectErr := pgxscan.Select(ctx, db, &player, `SELECT * from players`)
		if selectErr != nil {
			slog.Error("Error during select from database", selectErr)
			return
		}
		utils.CheckLastActionFinish(player, db)

		// create action
		duration := time.Duration(data.Options[1].IntValue())
		endAt := time.Now().Add(time.Second * duration)
		utils.AddAction(player.ID, "harvest | duration:"+data.Options[1].StringValue(), db, endAt)

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s", player),
			},
		})

	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		var choices []*discordgo.ApplicationCommandOptionChoice
		switch {
		case data.Options[0].Focused:
			var resourcesTypes []_struct.ResourcesType
			selectErr := pgxscan.Select(ctx, db, &resourcesTypes, `SELECT * FROM resources_types`)
			if selectErr != nil {
				slog.Error("Error during select from database", selectErr)
				return
			}
			for _, resourceType := range resourcesTypes {
				choice := discordgo.ApplicationCommandOptionChoice{
					Name:  resourceType.Name,
					Value: resourceType.ID,
				}
				choices = append(choices, &choice)
			}

		case data.Options[1].Focused:
			var resources []_struct.Resources
			selectErr := pgxscan.Select(ctx, db, &resources, `SELECT name, id FROM resources`)
			if selectErr != nil {
				slog.Error("Error during select from database", selectErr)
				return
			}

			timeChoices := []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "30min",
					Value: 1800,
				},
				{
					Name:  "1h",
					Value: 3600,
				},
				{
					Name:  "1h30",
					Value: 5400,
				},
				{
					Name:  "2h",
					Value: 7200,
				},
			}

			for _, choice := range timeChoices {
				choices = append(choices, choice)
			}

		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices,
			},
		})
		if err != nil {
			slog.Error("Error during AutoComplete Interaction Response", err)
			return
		}

	}

}
