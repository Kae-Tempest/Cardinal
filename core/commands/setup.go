package commands

import (
	_struct "Raphael/core/struct"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
)

func Setup(s *discordgo.Session, i *discordgo.InteractionCreate, db *pgxpool.Pool) {
	ctx := context.Background()

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()

		var player = _struct.Player{}

		raceID, raceErr := strconv.Atoi(data.Options[1].StringValue())
		jobID, jobErr := strconv.Atoi(data.Options[2].StringValue())

		if jobErr != nil || raceErr != nil {
			slog.Error("Error during parsing string to int", jobErr, raceErr)
		}

		player.Name = i.Interaction.Member.User.GlobalName
		player.ServerID = i.GuildID
		player.Username = data.Options[0].StringValue()
		player.RaceID = raceID
		player.JobID = jobID
		player.Exp = 0
		player.Po = 50
		player.Level = 1
		player.GuildID = 0

		parsePlayer, _ := json.Marshal(player)

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("User details is:%s", parsePlayer),
			},
		})
		if err != nil {
			slog.Error("Error during Interaction Response", err)
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		data := i.ApplicationCommandData()
		var choices []*discordgo.ApplicationCommandOptionChoice
		switch {
		case data.Options[1].Focused:
			var races []*_struct.Race
			selectErr := pgxscan.Select(ctx, db, &races, `SELECT * FROM public.races`)
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
			selectErr := pgxscan.Select(ctx, db, &jobs, `SELECT * FROM public.jobs`)
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
		}
	}
}
