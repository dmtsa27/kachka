package main

import (
	"context"
	"log"

	"github.com/dmtsa27/kachka.git/pkg/api"
	"github.com/dmtsa27/kachka.git/pkg/service"
	"github.com/dmtsa27/kachka.git/pkg/storage"
	"github.com/dmtsa27/kachka.git/pkg/telegram"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

// @title Kachka Bot API
// @version 1.0
// @description Admin API for Kachka Bot challenge management.
// @host localhost:8080
// @BasePath /
func main() {
	if err := godotenv.Overload(); err != nil {
		log.Println("No .env file found, using system variables")
	}

	app := fx.New(
		storage.Module,
		service.Module,
		telegram.Module,
		api.Module,
		fx.Invoke(func(lc fx.Lifecycle, bot *telegram.Bot, svc *service.Service) {
			bot.SetService(svc)
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						// Use Background context because fx OnStart context cancels after 15s
						if err := bot.Start(context.Background()); err != nil {
							log.Printf("Bot error: %v", err)
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					bot.Stop()
					return nil
				},
			})
		}),
	)

	app.Run()
}
