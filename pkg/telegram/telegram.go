package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Processor struct {
	bot    *tgbotapi.BotAPI
	offset int
}

func New(token string) (*Processor, error) {
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		return nil, err
	}

	return &Processor{
		bot: bot,
	}, nil
}
