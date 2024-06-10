package utils

import (
	_struct "Raphael/core/struct"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"math"
	"strings"
	"time"
)

func CheckLastActionFinish(player _struct.Player, db *pgxpool.Pool) {
	ctx := context.Background()

	// Get last user action

	var lastAction _struct.PlayerAction
	selectErr := pgxscan.Select(ctx, db, &lastAction, `SELECT * FROM players_actions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, player.ID)
	if selectErr != nil {
		slog.Error("Error during selecting in database", selectErr)
		return
	}

	// check is last action is duration Action

	if strings.Contains(lastAction.Action, "duration") || strings.Contains(lastAction.Action, "idle") {
		now := time.Now()

		if strings.Contains(lastAction.Action, "duration") || time.Time.Before(lastAction.EndAt, now) {
			duration := now.Sub(lastAction.CreatedAt)
			upsertResource(player, db, lastAction, ctx, duration)
		} else {
			duration := lastAction.EndAt.Sub(lastAction.CreatedAt)
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
			var resource _struct.Resources
			selectErr := pgxscan.Select(ctx, db, &resource, `SELECT id, name, quantities_per_min FROM resources`)
			if selectErr != nil {
				slog.Error("Error during selecting in database", selectErr)
				return
			}

			passedTime := math.Round(duration.Minutes() / 5)
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
