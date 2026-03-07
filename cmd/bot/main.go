package main

import (
	"context"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/dmtsa27/kachka.git/pkg/service"
	"github.com/dmtsa27/kachka.git/pkg/storage"
	"github.com/dmtsa27/kachka.git/pkg/telegram"
)

func main() {

	ctx := context.Background()

	if err := godotenv.Overload(); err != nil {
		log.Println("No .env file found, using system variables")
	}

	token := mustToken()
	dsn := os.Getenv("DATABASE_URL")
	log.Printf("DEBUG: Connecting to DSN: %s\n", dsn)

	if dsn == "" {
		log.Fatal("DSN empty")
	}

	mystorage, err := storage.NewPostgresDB(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to storage (PostgreSQL): %v", err)
	}

	log.Println("Connected to DB")

	bot, err := telegram.New(token, nil)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	svc := service.New(mystorage, nil)
	bot.SetService(svc)
	svc.SetNotifier(bot)

	if err := bot.Start(ctx); err != nil {
		log.Fatalf("Bot error: %v", err)
	}

}

func mustToken() string {

	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatal("invalid token")
	}

	return token

}
