package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/dmtsa27/kachka.git/pkg/storage"
)

func main() {

	ctx := context.Background()

	if err := godotenv.Overload(); err != nil {
		log.Println("No .env file found, using system variables")
	}

	_ = mustToken()
	dsn := os.Getenv("DATABASE_URL")
	fmt.Printf("DEBUG: Connecting to DSN: %s\n", dsn)

	if dsn == "" {
		log.Fatal("DSN empty")
	}

	mystorage, err := storage.NewPostgresDB(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to storage (PostgreSQL):  %v", err)
	}

	fmt.Println("Connected to DB")

}

func mustToken() string {

	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatal("invalid token")
	}

	return token

}
