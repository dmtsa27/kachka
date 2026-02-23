package telegram

import (
	"context"

	"github.com/dmtsa27/kachka.git/pkg/service"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Bot struct {
	client  *telego.Bot
	service *service.Service
}

func New(BotToken string, service *service.Service) (*Bot, error) {

	bot, err := telego.NewBot(BotToken)
	if err != nil {
		return nil, err
	}

	return &Bot{
		client:  bot,
		service: service,
	}, nil

}

func Start(ctx context.Context, bot *Bot) error {

	updates, err := bot.client.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return err
	}

	bh, err := th.NewBotHandler(bot.client, updates)

	if err != nil {
		return err
	}
	defer bh.Stop()

	bh.HandleCircle(func(ctx *th.Context, message telego.Message){
		if message.VideoNote == nil{
			return
		}
		
		
		

	}, th.AnyMessage()

}
