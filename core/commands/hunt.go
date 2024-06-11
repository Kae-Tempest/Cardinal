package commands

import (
	_struct "Raphael/core/struct"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strings"
)

func Hunt(s *discordgo.Session, i *discordgo.InteractionCreate, db *pgxpool.Pool) {
	ctx := context.Background()
	/*
	 * Selection Aleatoire de la creature celon l'emplacement et le niveau du joueur
	 * * Choix dans la table des apparition
	 * Début du combat -> Si dex joueur + haut que la creature le joueur commence sinon l'inverse
	 * Creation d'un fils privé avec le joueur et le bot afin de ne pas flood le chat
	 * Boucle While begin
	 * envoie de la demande de choix du skill au joueur -> Attaque de base + 3 skill définis pars le joueur
	 * en parralle choix automatique de l'attaque de la creature
	 * execution du tour en decontant les degats sur le joueur et la creature
	 * Si mort d'une des deux partie = Boucle While end
	 * Mort joueur : perte de x PO + retour en ville
	 * Mort Creature : Loot des Reward : { Po: X , Exp: X, items: [1,2,3] }
	 */
	// Get playet
	var player _struct.Player
	selectPlayerErr := pgxscan.Get(ctx, db, &player, `SELECT * FROM players where name = $1 LIMIT 1`, i.Interaction.Member.User.GlobalName)
	if selectPlayerErr != nil {
		slog.Error("Error during select into database", selectPlayerErr)
		return
	}

	// Selection de la creature
	//var creature _struct.Creatures

	// definition de l'ordre de jeux

	// get Channel fight ID
	channels, getChannelsErr := s.GuildChannels(i.GuildID)
	if getChannelsErr != nil {
		return
	}
	var parentID string
	for _, channel := range channels {
		if strings.Contains(channel.Name, "fights") && channel.Type == discordgo.ChannelTypeGuildText {
			parentID = channel.ID
		}
	}
	// creation du fils privé
	threadData := discordgo.ThreadStart{
		Name:                fmt.Sprintf("%s VS %s", player.Name, player.Name),
		AutoArchiveDuration: 60,
		Invitable:           false,
	}

	textMessage, textErr := s.ChannelMessageSend(parentID, "Hunt Begin..")
	if textErr != nil {
		slog.Error("Error during Thread Channel Creation", textErr)
		return
	}

	_, threadErr := s.MessageThreadStartComplex(parentID, textMessage.ID, &threadData)
	if threadErr != nil {
		slog.Error("Error during Thread Channel Creation", threadErr)
		return
	}
	// envoie du choix de skill

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong! 🏓",
		},
	})
}
