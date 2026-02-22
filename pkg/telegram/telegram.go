package telegram

import (
	"context"
	"log"

	"github.com/dmtsa27/kachka.git/pkg/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Processor struct {
	bot     *tgbotapi.BotAPI
	service *service.Service
}

func New(token string, svc *service.Service) (*Processor, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Processor{
		bot:     bot,
		service: svc,
	}, nil
}

// Start запускає polling і обробляє кожен апдейт
func (p *Processor) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := p.bot.GetUpdatesChan(u)

	for update := range updates {
		if err := p.Process(ctx, update); err != nil {
			log.Printf("process update: %v", err)
		}
	}
}

// Process обробляє один апдейт
func (p *Processor) Process(ctx context.Context, update tgbotapi.Update) error {
	// Кружок >= 30 сек → 👀 + HandleCircle
	if update.Message == nil {
        return nil
    }
	if update.Message.VideoNote != nil {
		duration := update.Message.VideoNote.Duration
		if duration >= 30 {
			p.sendReaction(update.Message.Chat.ID, update.Message.MessageID, "👀")

			return p.service.HandleCircle(
				ctx,
				update.Message.From.ID,
				duration,
				update.Message.Chat.ID,
				update.Message.MessageID,
			)
		}
	}

	// Реплай 🖕 на кружок → бот відповідає 🖕 + видаляє сесію
	if update.Message != nil && update.Message.ReplyToMessage != nil && update.Message.Text == "🖕" {
		replyTo := update.Message.ReplyToMessage
		if replyTo.VideoNote != nil {
			chatID := update.Message.Chat.ID
			msgID := replyTo.MessageID

			reply := tgbotapi.NewMessage(chatID, "🖕")
			reply.ReplyToMessageID = replyTo.MessageID
			p.bot.Send(reply)

			err := p.service.CancelSession(ctx, chatID, msgID)
			if err != nil {
				log.Printf("cancel session: %v", err)
			}
			return nil
		}
	}

	return nil
}

// sendReaction відправляє емодзі-реакцію на повідомлення
func (p *Processor) sendReaction(chatID int64, messageID int, emoji string) {
	params := tgbotapi.Params{}
	params.AddNonZero64("chat_id", chatID)
	params.AddNonZero("message_id", messageID)
	params["reaction"] = `[{"type":"emoji","emoji":"` + emoji + `"}]`
	params["is_big"] = "false"

	_, err := p.bot.MakeRequest("setMessageReaction", params)
	if err != nil {
		log.Printf("send reaction: %v", err)
	}
}
