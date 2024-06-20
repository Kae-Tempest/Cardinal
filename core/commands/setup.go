package commands

import (
	"Cardinal/core/entities"
	"Cardinal/core/rpg"
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

func Setup(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, db *pgxpool.Pool) {

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		var player entities.Player

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
		player.LocationId = 1

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
		_, insertErr := db.Exec(ctx, `INSERT into players (name, server_id, username, race_id, job_id, exp, po , level, guild_id, inventory_size, location_id) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 )`,
			player.Name, player.ServerID, player.Username, player.RaceID, player.JobID, player.Exp, player.Po, player.Level, player.GuildID, player.InventorySize, player.LocationId)
		if insertErr != nil {
			slog.Error("Error during insert from database", insertErr)
			return
		}

		var user *entities.Player
		selectErr := pgxscan.Get(ctx, db, &user, `SELECT id from players where name = $1 LIMIT 1`, player.Name)
		if selectErr != nil {
			slog.Error("Error during select from database", selectErr)
			return
		}

		// TODO: Get basic stats of Race or Job

		var job entities.Job
		jobSelectErr := pgxscan.Get(ctx, db, &job, `SELECT * from jobs where id = $1`, player.JobID)
		if jobSelectErr != nil {
			slog.Error("Error during select from database", jobSelectErr)
			return
		}

		var stats entities.Stats
		_, insertErr = db.Exec(ctx, `INSERT into stats (user_id, hp,  strength, constitution, mana, stamina, dexterity, intelligence, wisdom, charisma) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			user.ID, stats.HP, stats.Strength, stats.Constitution, stats.Mana, stats.Stamina, stats.Dexterity, stats.Intelligence, stats.Wisdom, stats.Charisma)

		rpg.AddAction(ctx, user.ID, "create character", db, time.Now())

	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		var choices []*discordgo.ApplicationCommandOptionChoice
		switch {
		case data.Options[1].Focused:
			var races []*entities.Race
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
			var jobs []*entities.Job
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
