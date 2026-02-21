package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/dmtsa27/kachka.git/pkg/storage"
)

func main() {

	if err := godotenv.Overload(); err != nil {
		log.Println("No .env file found, using system variables")
	}

	_ = mustToken()
	dsn := os.Getenv("DATABASE_URL")
	fmt.Printf("DEBUG: Connecting to DSN: %s\n", dsn)

	if dsn == "" {
		log.Fatal("DSN empty")
	}

	mystorage, err := storage.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to storage (PostgreSQL):  %v", err)
	}

	fmt.Println("Connected to DB")

	ctx := context.Background()
	myChallenge := storage.Challenge{
		IsActive:    true,
		DaysPerWeek: 3,
		Duration:    180,
	}
	err = mystorage.CreateChallenge(ctx, myChallenge)

	if err != nil {
		log.Fatalf("challenge could not be created: %v", err)
	}
	fmt.Println("Challenge created succesfully")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mystorage.GetChallenge(ctx, 1)
	if err != nil {
		fmt.Printf("Could not get challenge : %+v\n ", err)
	} else {
		fmt.Printf("Found challenge: %+v\n", c)
	}

}

func mustToken() string {

	token := os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatal("invalid token")
	}

	return token

}
