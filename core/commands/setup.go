package commands

import (
	_struct "Raphael/core/struct"
	"Raphael/core/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
	"time"
)

func Setup(s *discordgo.Session, i *discordgo.InteractionCreate, db *pgxpool.Pool) {
	ctx := context.Background()

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		var player _struct.Player

		raceID, _ := strconv.Atoi(data.Options[1].StringValue())
		jobID, _ := strconv.Atoi(data.Options[2].StringValue())

		player.Name = i.Interaction.Member.User.GlobalName
		player.ServerID = i.GuildID
		player.Username = data.Options[0].StringValue()
		player.RaceID = raceID
		player.JobID = jobID
		player.Exp = 0
		player.Po = 50
		player.Level = 1
		player.GuildID = 0
		player.InventorySize = 10

		parsePlayer, _ := json.Marshal(player)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("User details is:%s", parsePlayer),
			},
		})
		if err != nil {
			slog.Error("Error during Interaction Response", err)
			return
		}
		_, insertErr := db.Exec(ctx, `INSERT into players (name, server_id, username, race_id, job_id, exp, po , level, guild_id, inventory_size) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10 )`,
			player.Name, player.ServerID, player.Username, player.RaceID, player.JobID, player.Exp, player.Po, player.Level, player.GuildID, player.InventorySize)
		if insertErr != nil {
			slog.Error("Error during insert from database", insertErr)
			return
		}
		var user []*_struct.Player
		selectErr := pgxscan.Select(ctx, db, &user, `SELECT id from players where name = $1 LIMIT 1`, player.Name)
		if selectErr != nil {
			slog.Error("Error during select from database", selectErr)
			return
		}

		utils.AddAction(user[0].ID, "create character", db, time.Now())

	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		var choices []*discordgo.ApplicationCommandOptionChoice
		switch {
		case data.Options[1].Focused:
			var races []*_struct.Race
			selectErr := pgxscan.Select(ctx, db, &races, `SELECT * FROM races`)
			if selectErr != nil {
				slog.Error("Error during select from database", selectErr)
				return
			}

			for _, race := range races {
				choice := discordgo.ApplicationCommandOptionChoice{
					Name:  race.Name,
					Value: strconv.Itoa(race.ID),
				}
				choices = append(choices, &choice)
			}
		case data.Options[2].Focused:
			var jobs []*_struct.Job
			selectErr := pgxscan.Select(ctx, db, &jobs, `SELECT * FROM jobs`)
			if selectErr != nil {
				slog.Error("Error during select from database", selectErr)
				return
			}

			for _, job := range jobs {
				choice := discordgo.ApplicationCommandOptionChoice{
					Name:  job.Name,
					Value: strconv.Itoa(job.ID),
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
