package rpg

import (
	"Cardinal/core/entities"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func CreateHuntFightThead(s *discordgo.Session, i *discordgo.InteractionCreate, playerName string, creatureName string) *discordgo.Channel {
	channels, getChannelsErr := s.GuildChannels(i.GuildID)
	if getChannelsErr != nil {

	}
	var parentID string
	for _, channel := range channels {
		if strings.Contains(channel.Name, "fights") && channel.Type == discordgo.ChannelTypeGuildText {
			parentID = channel.ID
		}
	}
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

func HuntFight(s *discordgo.Session, player entities.Player, creature entities.Creatures, order []entities.FightOrder, threadChannel *discordgo.Channel, db *pgxpool.Pool) {
	ctx := context.Background()
	var playerSkill *skill
	var creatureSkill *skill
	var playerChosenSkill entities.Skill
	var creatureChosenSkill entities.Skill
	var playerStats entities.Stats
	var i int = 1

	getPlayerStatErr := pgxscan.Get(ctx, db, &playerStats, `SELECT * FROM stats where user_id = $1`, player.ID)
	if getPlayerStatErr != nil {
		slog.Error("Error during getting Player stats in database", getPlayerStatErr)
		return
	}
	for playerStats.HP > 0 || creature.HP > 0 {
		if order[0].Name == "Player" {
			playerSkill, creatureSkill = playerTurn(player, threadChannel, db, s), creatureTurn(creature, db)
			if creatureSkill == nil || playerSkill == nil {
				slog.Error("Error during choosing Skill")
				return
			}
			getSkillErr := pgxscan.Get(ctx, db, &playerChosenSkill, `SELECT * from skills where id = $1`, playerSkill.ID)
			getSkillErr = pgxscan.Get(ctx, db, &creatureChosenSkill, `SELECT * from skills where id = $1`, creatureSkill.ID)

			if getSkillErr != nil {
				slog.Error("Error during getting Skill in database", getSkillErr)
			}
		} else {
			creatureSkill, playerSkill = creatureTurn(creature, db), playerTurn(player, threadChannel, db, s)
			if creatureSkill == nil || playerSkill == nil {
				slog.Error("Error during choosing Skill")
				return
			}
			getSkillErr := pgxscan.Get(ctx, db, &playerChosenSkill, `SELECT * from skills where id = $1`, playerSkill.ID)
			getSkillErr = pgxscan.Get(ctx, db, &creatureChosenSkill, `SELECT * from skills where id = $1`, creatureSkill.ID)

			if getSkillErr != nil {
				slog.Error("Error during getting Skill in database", getSkillErr)
			}
		}

		if playerSkill.Type == creatureSkill.Type {
			if playerSkill.Type == "attack" {
				var damage int
				switch playerChosenSkill.Type {
				case "attack":
					damage = playerChosenSkill.Strength + playerStats.Strength
					creature.HP = creature.HP - damage
					break
				case "magic":
					damage = playerChosenSkill.Intelligence + playerStats.Intelligence
					creature.HP = creature.HP - damage
					break
				}
				playerStats.Mana -= playerChosenSkill.Mana

				_, sendErr := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("You did %d points of damage to %s", damage, creature.Name))
				if sendErr != nil {
					slog.Error("error during send message", sendErr)
				}
				switch creatureChosenSkill.Type {
				case "attack":
					damage = creatureChosenSkill.Strength + creature.Strength
					playerStats.HP = playerStats.HP - damage
					break
				case "magic":
					damage = creatureChosenSkill.Intelligence + creature.Intelligence
					playerStats.HP = playerStats.HP - damage
					break
				}
				creature.Mana -= creatureChosenSkill.Mana
				// send msg for HP update
				_, sendErr = s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("%s did %d points of damage to you", creature.Name, damage))
				if sendErr != nil {
					slog.Error("error during send message", sendErr)
				}

			} else {
				_, err := s.ChannelMessageSend(threadChannel.ID, "Nothing append this turn...")
				if err != nil {
					slog.Error("Error during sending message", err)
					return
				}
			}
		} else if playerSkill.Type == "attack" && creatureSkill.Type == "block" {
			var damage int
			switch playerChosenSkill.Type {
			case "attack":
				damage = (playerChosenSkill.Strength + playerStats.Strength) - creature.Constitution
				if damage < 0 {
					damage = 0
				}
				creature.HP = creature.HP - damage
				break
			case "magic":
				damage = (playerChosenSkill.Intelligence + playerStats.Intelligence) - creature.Constitution
				if damage < 0 {
					damage = 0
				}
				creature.HP = creature.HP - damage
				break
			}
			playerStats.Mana -= playerChosenSkill.Mana
			_, sendErr := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("You did %d points of damage to %s", damage, creature.Name))
			if sendErr != nil {
				slog.Error("error during send message", sendErr)
			}
		} else if playerSkill.Type == "attack" && creatureSkill.Type == "dodge" {
			if playerStats.Dexterity > creature.Dexterity {
				var damage int
				switch playerChosenSkill.Type {
				case "attack":
					damage = (playerChosenSkill.Strength + playerStats.Strength) * (5 / 100)
					if damage < 0 {
						damage = 0
					}
					creature.HP = creature.HP - damage
					break
				case "magic":
					damage = (playerChosenSkill.Intelligence + playerStats.Intelligence) * (5 / 100)
					if damage < 0 {
						damage = 0
					}
					creature.HP = creature.HP - damage
					break
				}
				_, sendErr := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("You did %d points of damage to %s", damage, creature.Name))
				if sendErr != nil {
					slog.Error("error during send message", sendErr)
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
			var damage int
			switch creatureChosenSkill.Type {
			case "attack":
				damage = (creatureChosenSkill.Strength + creature.Strength) - playerStats.Constitution
				if damage < 0 {
					damage = 0
				}
				playerStats.HP = playerStats.HP - damage
				break
			case "magic":
				damage = (creatureChosenSkill.Intelligence + creature.Intelligence) - playerStats.Constitution
				if damage < 0 {
					damage = 0
				}
				playerStats.HP = playerStats.HP - damage
				break
			}
			playerStats.Mana -= playerChosenSkill.Mana
			// Send msg for HP update
			_, sendErr := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("You did %d points of damage to %s", damage, creature.Name))
			if sendErr != nil {
				slog.Error("error during send message", sendErr)
			}
		} else if playerSkill.Type == "dodge" && creatureSkill.Type == "attack" {
			if playerStats.Dexterity < creature.Dexterity {
				var damage int
				switch creatureChosenSkill.Type {
				case "attack":
					damage = (creatureChosenSkill.Strength + creature.Strength) * (5 / 100)
					if damage < 0 {
						damage = 0
					}
					playerStats.HP = playerStats.HP - damage
					break
				case "magic":
					damage = (creatureChosenSkill.Intelligence + creature.Intelligence) * (5 / 100)
					if damage < 0 {
						damage = 0
					}
					playerStats.HP = playerStats.HP - damage
					break
				}
				_, sendErr := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("%s did %d points of damage to you", creature.Name, damage))
				if sendErr != nil {
					slog.Error("error during send message", sendErr)
				}
			} else {
				_, err := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("You have dodge %s attack...", creature.Name))
				if err != nil {
					slog.Error("Error during sending message", err)
					return
				}
			}
			creature.Mana -= creatureChosenSkill.Mana
		} else if playerSkill.Type == "attack" && creatureSkill.Type == "dodge" {
			var damage int
			if playerStats.Dexterity > creature.Dexterity {
				switch creatureChosenSkill.Type {
				case "attack":
					damage = (playerChosenSkill.Strength + playerStats.Strength) * (5 / 100)
					if damage < 0 {
						damage = 0
					}
					creature.HP = creature.HP - damage
					break
				case "magic":
					damage = (playerChosenSkill.Intelligence + playerStats.Intelligence) * (5 / 100)
					if damage < 0 {
						damage = 0
					}
					creature.HP = creature.HP - damage
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
		} else {
			_, err := s.ChannelMessageSend(threadChannel.ID, "Nothing append this turn...")
			if err != nil {
				slog.Error("Error during sending message", err)
				return
			}
		}
	}
	_, sendErr := s.ChannelMessageSend(threadChannel.ID, fmt.Sprintf("End of turn %d", i))
	if sendErr != nil {
		slog.Error("Error during sending message", sendErr)
	}
	i++
	// while end
}

func creatureTurn(creature entities.Creatures, db *pgxpool.Pool) *skill {
	var basicCreatureSkill []*skill

	basicCreatureSkill = append(basicCreatureSkill, &skill{
		Type: "attack",
		ID:   1,
		Name: "Basic attack",
	})
	basicCreatureSkill = append(basicCreatureSkill, &skill{
		Type: "block",
		ID:   2,
		Name: "Basic block",
	})
	basicCreatureSkill = append(basicCreatureSkill, &skill{
		Type: "dodge",
		ID:   3,
		Name: "Basic dodge",
	})

	// Choix d'une attaque aleatoire | Monstre < lvl 15 = attaque de base, esquive , blocage
	if creature.Level < 15 {
		r := rand.Intn(10)
		for _, skill := range basicCreatureSkill {
			switch {
			case r <= 5 && skill.Type == "attack":
				return skill
			case r > 6 && r < 8 && skill.Type == "block":
				return skill
			case r >= 8 && skill.Type == "dodge":
				return skill
			}
		}
	} else if creature.Level >= 15 {
		// get random skill of selected monster
		var creatureSkills []entities.Skill
		selectErr := pgxscan.Select(context.Background(), db, &creatureSkills, `select id, name from skills s join creature_skill cs on s.id = cs.skill_id where cs.creature_id = $1`, creature.ID)
		if selectErr != nil {
			slog.Error("Error during selection creature Skills into databases from creature ID", selectErr)
			return nil
		}

		for _, cSkill := range creatureSkills {
			basicCreatureSkill = append(basicCreatureSkill, &skill{
				Type: cSkill.Type,
				ID:   cSkill.ID,
				Name: cSkill.Name,
			})
		}

		r := rand.Intn(10)
		for _, skill := range basicCreatureSkill {
			switch {
			case r <= 5 && r > 3 && skill.Name == "attack":
				return skill
			case r > 6 && r < 8 && skill.Name == "block":
				return skill
			case r >= 8 && skill.Name == "dodge":
				return skill
			default:
				return skill
			}
		}
	}
	// renvoie du skill choisie
	return nil
}

func playerTurn(player entities.Player, threadChannel *discordgo.Channel, db *pgxpool.Pool, s *discordgo.Session) *skill {
	var basicPlayerSkill []*skill

	basicPlayerSkill = append(basicPlayerSkill, &skill{
		Type: "attack",
		ID:   1,
		Name: "Basic attack",
	})
	basicPlayerSkill = append(basicPlayerSkill, &skill{
		Type: "block",
		ID:   2,
		Name: "Basic block",
	})
	basicPlayerSkill = append(basicPlayerSkill, &skill{
		Type: "dodge",
		ID:   3,
		Name: "Basic dodge",
	})

	// Get 3 skill + atk de base + dodge + block
	var playerSkills []entities.Skill
	selectErr := pgxscan.Select(context.Background(), db, &playerSkills, `select id, name from skills s join user_skill us on s.id = us.skill_id join user_job_skill ujs on s.id = ujs.job_skill_id where us.user_id = $1 and ujs.user_id = $1`, player.ID)
	if selectErr != nil {
		slog.Error("Error during selection player Skills into databases", selectErr)
	}

	// creer un select btn pour les skills | creer les btn pour les basicSkill

	btnAtk := discordgo.Button{
		Label:    "Basic attack",
		Style:    discordgo.PrimaryButton,
		CustomID: "attack",
	}
	btnblk := discordgo.Button{
		Label:    "Basic block",
		Style:    discordgo.PrimaryButton,
		CustomID: "block",
	}
	btnDdg := discordgo.Button{
		Label:    "Basic dodge",
		Style:    discordgo.PrimaryButton,
		CustomID: "dodge",
	}

	var skillOptions []discordgo.SelectMenuOption
	if len(playerSkills) > 0 {
		for _, playerSkill := range playerSkills {
			skillOptions = append(skillOptions, discordgo.SelectMenuOption{
				Label:       playerSkill.Name,
				Value:       strconv.Itoa(playerSkill.ID),
				Description: playerSkill.Description,
				Default:     false,
			})
		}
	}
	var selectbtn discordgo.SelectMenu
	if len(skillOptions) > 0 {
		selectbtn = discordgo.SelectMenu{
			MenuType:    discordgo.StringSelectMenu,
			CustomID:    "skillselectbtn",
			Placeholder: "You're Skills",
			Options:     skillOptions,
		}
	} else {
		selectbtn = discordgo.SelectMenu{
			MenuType:    discordgo.StringSelectMenu,
			CustomID:    "skillselectbtn",
			Placeholder: "You're Skills",
			Disabled:    true,
			Options: []discordgo.SelectMenuOption{
				{
					Label:       "Any",
					Value:       "0",
					Description: "any",
					Default:     false,
				},
			},
		}
	}

	messageData := discordgo.MessageSend{
		Content: "Choose you're skill",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					btnAtk, btnDdg, btnblk,
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					selectbtn,
				},
			},
		},
	}
	// envoyer le messages des skills

	message, err := s.ChannelMessageSendComplex(threadChannel.ID, &messageData)
	if err != nil {
		slog.Error("Error during sending the message", err)
	}
	// reception du choix du joueur
	result, buttonInteractionErr := waitForButtonInteraction(s, message.ID, threadChannel.ID, 5*time.Minute)
	if buttonInteractionErr != nil {
		slog.Error("Error durring wait interraction", buttonInteractionErr)
	}
	// renvoie du skill choisie

	switch result {
	case "attack":
		return basicPlayerSkill[0]
	case "block":
		return basicPlayerSkill[1]
	case "dodge":
		return basicPlayerSkill[2]

	}
	return nil
}

// waitForButtonInteraction waits for a button interaction and returns the CustomID of the clicked button.
func waitForButtonInteraction(s *discordgo.Session, messageID, channelID string, timeout time.Duration) (string, error) {
	interactionChannel := make(chan *discordgo.InteractionCreate)
	defer close(interactionChannel)

	// Temporary handler for this interaction
	tempHandler := func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if interaction.Message.ID == messageID && interaction.ChannelID == channelID {
			interactionChannel <- interaction
		}
	}

	// Add the temporary handler
	s.AddHandler(tempHandler)

	// Create a wait group to handle timeout
	var wg sync.WaitGroup
	wg.Add(1)

	var result *discordgo.InteractionCreate
	var err error

	go func() {
		defer wg.Done()
		select {
		case result = <-interactionChannel:
		case <-time.After(timeout):
			err = fmt.Errorf("timeout waiting for interaction")
		}
	}()

	// Wait for the interaction or timeout
	wg.Wait()

	if err != nil {
		return "", err
	}

	_ = s.InteractionRespond(result.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "",
			Components: []discordgo.MessageComponent{},
		},
	})

	return result.MessageComponentData().CustomID, nil
}
