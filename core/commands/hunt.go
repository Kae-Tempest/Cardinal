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
	"math/rand"
)

func Hunt(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, db *pgxpool.Pool) {
	/*
	 * Boucle While begin
	 * envoie de la demande de choix du skill au joueur -> Attaque de base + 3 skill définis pars le joueur
	 * en parralle choix automatique de l'attaque de la creature
	 * execution du tour en decontant les degats sur le joueur et la creature
	 * Si mort d'une des deux partie = Boucle While end
	 * Mort joueur : perte de x PO + retour en ville
	 * Mort Creature : Loot des Reward : { Po: X , Exp: X, items: [1,2,3] }
	 */

	// Get player
	var player entities.Player
	selectPlayerErr := pgxscan.Get(ctx, db, &player, `SELECT * FROM players where name = $1 LIMIT 1`, i.Interaction.Member.User.GlobalName)
	if selectPlayerErr != nil {
		slog.Error("Error during select player into database", selectPlayerErr)
		return
	}
	var playerStats entities.Stats
	selectPlayerStatErr := pgxscan.Get(ctx, db, &playerStats, `SELECT * FROM stats where user_id = $1 LIMIT 1`, player.ID)
	if selectPlayerStatErr != nil {
		slog.Error("Error during select player stats into database", selectPlayerErr)
		return
	}

	// Selection de la creature
	var locationCreature []entities.CreatureSpawns
	creaturesGetErr := pgxscan.Select(ctx, db, &locationCreature, `SELECT * FROM creaturespawn where emplacement_id = $1`, player.LocationId)
	if creaturesGetErr != nil {
		slog.Error("Error during select Creature's location into database", creaturesGetErr)
		return
	}

	selectedCreatureID := locationCreature[rand.Intn(len(locationCreature))].CreatureID
	var creature entities.Creatures
	creatureGetErr := pgxscan.Get(ctx, db, &creature, `SELECT * FROM creatures where id = $1`, selectedCreatureID)
	if creatureGetErr != nil {
		slog.Error("Error during selecting Creature into database", creatureGetErr)
		return
	}
	// definition de l'ordre de jeux

	var order []entities.FightOrder
	p := entities.FightOrder{
		Name: "Player",
		ID:   player.ID,
	}
	c := entities.FightOrder{
		Name: "Creature",
		ID:   creature.ID,
	}

	if playerStats.Dexterity > creature.Dexterity {

		order = append(order, p)
		order = append(order, c)
	} else {
		order = append(order, c)
		order = append(order, p)
	}

	threadChannel := rpg.CreateHuntFightThead(s, i, player.Username, creature.Name)
	// Boucle while
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Go to.. <#%s>", threadChannel.ID),
		},
	})

	if err != nil {
		slog.Error("Error during sending Interaction Respond", err)
		return
	}
	rpg.HuntFight(s, player, creature, order, threadChannel, db)

	// envoie du choix de skill

}
