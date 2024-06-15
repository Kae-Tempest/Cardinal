package utils

import (
	_struct "Raphael/core/struct"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"math"
	"math/rand"
	"strings"
	"time"
)

func CheckLastActionFinish(player _struct.Player, db *pgxpool.Pool) {
	ctx := context.Background()

	// Get last user action

	var lastAction _struct.PlayerAction
	selectErr := pgxscan.Get(ctx, db, &lastAction, `SELECT * FROM players_actions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, player.ID)
	if selectErr != nil {
		slog.Error("Error during selecting in database", selectErr)
		return
	}

	// check is last action is duration Action

	if strings.Contains(lastAction.Action, "duration") || strings.Contains(lastAction.Action, "idle") {
		now := time.Now()

		if strings.Contains(lastAction.Action, "duration") || time.Time.Before(lastAction.EndAt, now) {
			duration := lastAction.CreatedAt.Sub(now)
			upsertResource(player, db, lastAction, ctx, duration)
		} else {
			duration := lastAction.CreatedAt.Sub(lastAction.EndAt)
			upsertResource(player, db, lastAction, ctx, duration)
		}

	}
}

func AddAction(id int, actionName string, db *pgxpool.Pool, endAt time.Time) {

	_, insertErr := db.Exec(context.Background(), `INSERT into players_actions values ($1, $2, $3, $4)`, id, actionName, time.Now().Format("02_01_2006 15:04:05 -07:00"), endAt.Format("02_01_2006 15:04:05 -07:00"))
	if insertErr != nil {
		slog.Error("Error during insert action in database", insertErr)
		return
	}
}

func upsertResource(player _struct.Player, db *pgxpool.Pool, action _struct.PlayerAction, ctx context.Context, duration time.Duration) {
	var resourceTypes []_struct.ResourcesType
	selectErr := pgxscan.Select(ctx, db, &resourceTypes, `SELECT * FROM resources_types`)
	if selectErr != nil {
		slog.Error("Error during selecting in database", selectErr)
		return
	}
	for _, resourceType := range resourceTypes {
		if strings.Contains(action.Action, resourceType.Name) {
			fmt.Println(resourceType.Name, "Name")
			var resource _struct.Resources
			selectErr := pgxscan.Get(ctx, db, &resource, `SELECT id, name, quantities_per_min FROM resources where resources_type_id = $1`, resourceType.ID)
			if selectErr != nil {
				slog.Error("Error during selecting in database", selectErr)
				return
			}
			fmt.Println(duration)
			passedTime := math.Round(duration.Minutes() / 5)
			fmt.Println(passedTime)
			gatheredResources := int(passedTime) * resource.QuantitiesPerMin
			_, upsertError := db.Exec(ctx, `INSERT INTO ressourceinventory (user_id, item_id, quantity) values ($1,$2,$3) on CONFLICT(item_id)
					DO UPDATE SET quantity = excluded.quantity + ressourceinventory.quantity where ressourceinventory.user_id = excluded.user_id;`, player.ID, resource.ID, gatheredResources)
			if upsertError != nil {
				slog.Error("Error during upsert in database", upsertError)
				return
			}
		}
	}
}

func CreateHuntFightThead(s *discordgo.Session, i *discordgo.InteractionCreate, playerName string, creatureName string) *discordgo.Channel {
	// get Channel fight ID
	channels, getChannelsErr := s.GuildChannels(i.GuildID)
	if getChannelsErr != nil {

	}
	var parentID string
	for _, channel := range channels {
		if strings.Contains(channel.Name, "fights") && channel.Type == discordgo.ChannelTypeGuildText {
			parentID = channel.ID
		}
	}
	// creation du fils privé
	threadData := discordgo.ThreadStart{
		Name:                fmt.Sprintf("%s VS %s", playerName, creatureName),
		AutoArchiveDuration: 60,
		Invitable:           false,
	}

	textMessage, textErr := s.ChannelMessageSend(parentID, "Hunt Begin..")
	if textErr != nil {
		slog.Error("Error during Thread Channel Creation", textErr)
	}

	threadChannel, threadErr := s.MessageThreadStartComplex(parentID, textMessage.ID, &threadData)
	if threadErr != nil {
		slog.Error("Error during Thread Channel Creation", threadErr)
	}

	return threadChannel
}

type skill struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func HuntFight(s *discordgo.Session, player _struct.Player, creature _struct.Creatures, order []_struct.FightOrder, threadChannel *discordgo.Channel, db *pgxpool.Pool) {
	ctx := context.Background()
	var playerSkill skill
	var creatureSkill skill
	var playerChosenSkill _struct.Skill
	var creatureChosenSkill _struct.Skill
	var playerStats _struct.Stats

	getPlayerStatErr := pgxscan.Get(ctx, db, &playerStats, `SELECT * FROM stats where user_id = $1`, player.ID)
	if getPlayerStatErr != nil {
		slog.Error("Error during getting Player stats in database", getPlayerStatErr)
		return
	}

	if order[0].Name == "Player" {
		playerSkill, creatureSkill = playerTurn(player, threadChannel), creatureTurn(creature)
		getSkillErr := pgxscan.Get(ctx, db, &playerChosenSkill, `SELECT * from skills where name = $1`, playerSkill.Name)
		getSkillErr = pgxscan.Get(ctx, db, &creatureChosenSkill, `SELECT * from skills where name = $1`, creatureSkill.Name)

		if getSkillErr != nil {
			slog.Error("Error during getting Skill in database", getSkillErr)
		}
	} else {
		creatureSkill, playerSkill = creatureTurn(creature), playerTurn(player, threadChannel)
		getSkillErr := pgxscan.Get(ctx, db, &playerChosenSkill, `SELECT * from skills where name = $1`, playerSkill.Name)
		getSkillErr = pgxscan.Get(ctx, db, &creatureChosenSkill, `SELECT * from skills where name = $1`, creatureSkill.Name)

		if getSkillErr != nil {
			slog.Error("Error during getting Skill in database", getSkillErr)
		}
	}

	if playerSkill.Type == creatureSkill.Type {
		if playerSkill.Type == "attack" {
			switch playerChosenSkill.Type {
			case "attack":
				creature.HP = creature.HP - (playerChosenSkill.Strength + playerStats.Strength)
				break
			case "magic":
				creature.HP = creature.HP - (playerChosenSkill.Intelligence + playerStats.Intelligence)
				break
			}
			playerStats.Mana -= playerChosenSkill.Mana
			switch creatureChosenSkill.Type {
			case "attack":
				playerStats.HP = playerStats.HP - (creatureChosenSkill.Intelligence + creature.Intelligence)
				break
			case "magic":
				playerStats.HP = playerStats.HP - (creatureChosenSkill.Intelligence + creature.Intelligence)
				break
			}
			creature.Mana -= creatureChosenSkill.Mana
		} else {
			_, err := s.ChannelMessageSend(threadChannel.ID, "Nothing append this turn...")
			if err != nil {
				slog.Error("Error during sending message", err)
				return
			}
		}
	} else if playerSkill.Type == "attack" && creatureSkill.Type == "block" {
		switch playerChosenSkill.Type {
		case "attack":
			creature.HP = creature.HP - ((playerChosenSkill.Strength + playerStats.Strength) - creature.Constitution)
			break
		case "magic":
			creature.HP = creature.HP - ((playerChosenSkill.Intelligence + playerStats.Intelligence) - creature.Constitution)
			break
		}
		playerStats.Mana -= playerChosenSkill.Mana
	} else if playerSkill.Type == "attack" && creatureSkill.Type == "dodge" {
		if playerStats.Dexterity > creature.Dexterity {
			switch playerChosenSkill.Type {
			case "attack":
				creature.HP = creature.HP - ((playerChosenSkill.Strength + playerStats.Strength) * (5 / 100))
				break
			case "magic":
				creature.HP = creature.HP - ((playerChosenSkill.Intelligence + playerStats.Intelligence) * (5 / 100))
				break
			}
		} else {
			_, err := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("%s have dodge you're attack...", creature.Name))
			if err != nil {
				slog.Error("Error during sending message", err)
				return
			}
		}
		playerStats.Mana -= playerChosenSkill.Mana
	} else if playerSkill.Type == "block" && creatureSkill.Type == "attack" {
		switch creatureChosenSkill.Type {
		case "attack":
			damage := (creatureChosenSkill.Strength + creature.Strength) - playerStats.Constitution
			if damage < 0 {
				damage = 0
			}
			playerStats.HP = playerStats.HP - damage
			break
		case "magic":
			damage := (creatureChosenSkill.Intelligence + creature.Intelligence) - playerStats.Constitution
			if damage < 0 {
				damage = 0
			}
			playerStats.HP = playerStats.HP - damage
			break
		}
		playerStats.Mana -= playerChosenSkill.Mana
	} else if playerSkill.Type == "dodge" && creatureSkill.Type == "attack" {
		if playerStats.Dexterity < creature.Dexterity {
			switch creatureChosenSkill.Type {
			case "attack":
				playerStats.HP = playerStats.HP - ((creatureChosenSkill.Strength + creature.Strength) * (5 / 100))
				break
			case "magic":
				playerStats.HP = playerStats.HP - ((creatureChosenSkill.Intelligence + creature.Intelligence) * (5 / 100))
				break
			}
		} else {
			_, err := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("You have dodge %s attack...", creature.Name))
			if err != nil {
				slog.Error("Error during sending message", err)
				return
			}
		}
		creature.Mana -= creatureChosenSkill.Mana
	} else {
		_, err := s.ChannelMessageSend(threadChannel.ID, "Nothing append this turn...")
		if err != nil {
			slog.Error("Error during sending message", err)
			return
		}
	}
}

func creatureTurn(creature _struct.Creatures) skill {
	// Choix d'une attaque aleatoire | Monstre < lvl 15 = attaque de base, esquive , blocage
	if creature.Level < 15 {
		r := rand.Intn(10)
		switch {
		case r <= 5:
			return skill{
				Type: "attack",
				ID:   0,
				Name: "Basic attack",
			}
		case r > 6 && r < 8:
			return skill{
				Type: "block",
				ID:   0,
				Name: "Basic block",
			}
		case r >= 8:
			return skill{
				Type: "dodge",
				ID:   0,
				Name: "Basic dodge",
			}

		}
	} else {
		// get random skill of selected monster
	}
	// renvoie du skill choisie
	return skill{}
}

func playerTurn(player _struct.Player, threadChannel *discordgo.Channel) skill {
	// Get 3 skill + atk de base + dodge + block

	// envoyer le messages des skills
	// reception du choix du joueur
	// renvoie du skill choisie
	return skill{}
}
