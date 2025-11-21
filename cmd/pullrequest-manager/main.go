package main

import (
	"context"
	"log"
	"os"
	"pullrequest-inator/internal/application/services"
	"pullrequest-inator/internal/infrastructure/database/pg"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	} else {
		port = ":" + port
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to connect to database:", err)
	}
	defer pool.Close()

	prRepo := pg.NewPullRequestRepository(pool)
	statusRepo := pg.NewStatusRepository(pool)
	teamRepo := pg.NewTeamRepository(pool)
	userRepo := pg.NewUserRepository(pool)

	teamService, err := services.NewDefaultPullRequestService(userRepo, prRepo, teamRepo, statusRepo)
	_ = teamService

}
