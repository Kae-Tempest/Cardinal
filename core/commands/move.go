package commands

import (
	"Raphael/core/rpg"
	_struct "Raphael/core/struct"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
	"time"
)

func Move(s *discordgo.Session, i *discordgo.InteractionCreate, db *pgxpool.Pool) {
	ctx := context.Background()
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		var player _struct.Player
		selectErr := pgxscan.Get(ctx, db, &player, `SELECT * from players where name = $1`, i.Interaction.Member.User.GlobalName)
		if selectErr != nil {
			slog.Error("Error during select from database", selectErr)
			return
		}
		rpg.CheckLastActionFinish(player, db)

		locationID, _ := strconv.Atoi(data.Options[0].StringValue())
		var locationName string
		err := pgxscan.Get(ctx, db, &locationName, `SELECT name from locations where id = $1`, locationID)
		if err != nil {
			return
		}

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("%s moved to %s", player.Username, locationName),
			},
		})

		// update player location
		_, updateErr := db.Exec(ctx, `UPDATE players SET location_id = $1 where players.id = $2`,
			locationID, player.ID)
		if updateErr != nil {
			slog.Error("Error during update from database", updateErr)
			return
		}
		// insert player action
		rpg.AddAction(player.ID, "move to "+locationName, db, time.Now())

	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		var choices []*discordgo.ApplicationCommandOptionChoice
		switch {
		case data.Options[0].Focused:
			var locations []_struct.Locations
			selectErr := pgxscan.Select(ctx, db, &locations, `SELECT name, id FROM locations`)
			if selectErr != nil {
				slog.Error("Error during select from database", selectErr)
				return
			}

			for _, location := range locations {
				choice := discordgo.ApplicationCommandOptionChoice{
					Name:  location.Name,
					Value: strconv.Itoa(location.ID),
				}
				choices = append(choices, &choice)
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
