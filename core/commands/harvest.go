package commands

import (
	"Cardinal/core/entities"
	"Cardinal/core/rpg"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
	"time"
)

func Harvest(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, db *pgxpool.Pool) {

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		var player entities.Player
		selectErr := pgxscan.Get(ctx, db, &player, `SELECT * from players where name = $1`, i.Interaction.Member.User.GlobalName)
		if selectErr != nil {
			slog.Error("Error during select from database", selectErr)
			return
		}
		rpg.CheckLastActionFinish(ctx, player, db)

		// create action
		resourceChoice, _ := strconv.Atoi(data.Options[0].StringValue())
		durationChoice, _ := strconv.Atoi(data.Options[1].StringValue())
		var resource entities.Resources
		duration := time.Duration(durationChoice)
		err := pgxscan.Get(ctx, db, &resource, `SELECT name, id FROM resources_types where id = $1`, resourceChoice)
		if err != nil {
			return
		}
		endAt := time.Now().Add(time.Second * duration)
		rpg.AddAction(ctx, player.ID, "harvest"+resource.Name+"| duration:"+data.Options[1].StringValue(), db, endAt)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s begin to harvest %s for %s secondes", player.Username, resource.Name, data.Options[1].StringValue()),
			},
		})

	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		var choices []*discordgo.ApplicationCommandOptionChoice
		switch {
		case data.Options[0].Focused:
			var resourcesTypes []entities.ResourcesType
			selectErr := pgxscan.Select(ctx, db, &resourcesTypes, `SELECT * FROM resources_types`)
			if selectErr != nil {
				slog.Error("Error during select from database", selectErr)
				return
			}
			for _, resourceType := range resourcesTypes {
				choice := discordgo.ApplicationCommandOptionChoice{
					Name:  resourceType.Name,
					Value: strconv.Itoa(resourceType.ID),
				}
				choices = append(choices, &choice)
			}

		case data.Options[1].Focused:
			var resources []entities.Resources
			selectErr := pgxscan.Select(ctx, db, &resources, `SELECT name, id FROM resources`)
			if selectErr != nil {
				slog.Error("Error during select from database", selectErr)
				return
			}

			// TODO: Put data in DB ?
			timeChoices := []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "30min",
					Value: "1800",
				},
				{
					Name:  "1h",
					Value: "3600",
				},
				{
					Name:  "1h30",
					Value: "5400",
				},
				{
					Name:  "2h",
					Value: "7200",
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
