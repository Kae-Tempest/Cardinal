package Cardinal

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func DatabaseConnect() *pgxpool.Pool {
	ctx := context.Background()
	db, _ := pgxpool.New(ctx, os.Getenv("DB_URL"))

	return db
}
