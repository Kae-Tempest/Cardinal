package rpg

import (
	"Cardinal/core/entities"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"math"
	"strings"
	"time"
)

func CheckLastActionFinish(ctx context.Context, player entities.Player, db *pgxpool.Pool) {

	// Get last user action

	var lastAction entities.PlayerAction
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
			upsertResource(ctx, player, db, lastAction, duration)
		} else {
			duration := lastAction.CreatedAt.Sub(lastAction.EndAt)
			upsertResource(ctx, player, db, lastAction, duration)
		}

	}
}

func AddAction(ctx context.Context, playerID int, actionName string, db *pgxpool.Pool, startAt time.Time, endAt time.Time) {

	_, insertErr := db.Exec(ctx, `INSERT into players_actions (user_id, action, created_at, end_at) values ($1, $2, $3, $4)`, playerID, actionName, startAt.Format("02_01_2006 15:04:05 -07:00"), endAt.Format("02_01_2006 15:04:05 -07:00"))
	if insertErr != nil {
		slog.Error("Error during insert action in database", insertErr)
		return
	}
}

func upsertResource(ctx context.Context, player entities.Player, db *pgxpool.Pool, action entities.PlayerAction, duration time.Duration) {
	var resourceTypes []entities.ResourcesType
	selectErr := pgxscan.Select(ctx, db, &resourceTypes, `SELECT * FROM resources_types`)
	if selectErr != nil {
		slog.Error("Error during selecting in database", selectErr)
		return
	}
	for _, resourceType := range resourceTypes {
		if strings.Contains(action.Action, resourceType.Name) {
			var resource entities.Resources
			selectErr := pgxscan.Get(ctx, db, &resource, `SELECT id, name, quantities_per_min FROM resources where resources_type_id = $1`, resourceType.ID)
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
