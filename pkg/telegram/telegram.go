package telegram

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/service"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type BotService interface {
	RegisterUser(ctx context.Context, telegramID int64, username string) error
	HandleCircle(ctx context.Context, userID int64, duration int, chatID int64, messageID int) error
	StartChallenge(ctx context.Context, chatID int64, daysPerWeek int, duration int) error
	DeactivateChallengeForChat(ctx context.Context, chatID int64) error
	WeeklyCheck(ctx context.Context) ([]service.UserInfo, error)
	ActiveChallengeInfo(ctx context.Context) (chatID int64, nextCheck time.Time, err error)
}

type Bot struct {
	client  *telego.Bot
	service BotService
}

func (bot *Bot) SetService(svc BotService) {
	bot.service = svc
}

func (bot *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   text,
	})
	return err
}

func New(token string, svc BotService) (*Bot, error) {
	bot, err := telego.NewBot(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		client:  bot,
		service: svc,
	}, nil
}

func (bot *Bot) Start(ctx context.Context) error {
	updates, err := bot.client.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return err
	}

	bh, err := th.NewBotHandler(bot.client, updates)
	if err != nil {
		return err
	}
	defer bh.Stop()

	// VideoNote (circle) messages → try to count as workout
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		if message.VideoNote == nil || message.From == nil {
			return nil
		}

		userID := message.From.ID
		duration := message.VideoNote.Duration
		chatID := message.Chat.ID
		messageID := message.MessageID

		if err := bot.service.HandleCircle(ctx, userID, duration, chatID, messageID); err != nil {
			log.Printf("HandleCircle error (user=%d): %v", userID, err)
		}
		return nil
	}, th.AnyMessage())

	// Reactions → register user
	bh.HandleMessageReaction(func(ctx *th.Context, reaction telego.MessageReactionUpdated) error {
		if reaction.User == nil {
			return nil
		}

		userID := reaction.User.ID
		username := reaction.User.Username

		if err := bot.service.RegisterUser(ctx, userID, username); err != nil {
			log.Printf("RegisterUser error (user=%d): %v", userID, err)
		}
		return nil
	})

	// Bot доданий в групу → привітальне повідомлення + старт челенджу.
	// Бот кікнутий/чат видалений → деактивуємо челендж.
	bh.HandleMyChatMemberUpdated(func(ctx *th.Context, update telego.ChatMemberUpdated) error {
		oldStatus := update.OldChatMember.MemberStatus()
		newStatus := update.NewChatMember.MemberStatus()
		chatID := update.Chat.ID

		isJoining := (oldStatus == telego.MemberStatusLeft || oldStatus == telego.MemberStatusBanned) &&
			(newStatus == telego.MemberStatusMember || newStatus == telego.MemberStatusAdministrator)

		isLeaving := (oldStatus == telego.MemberStatusMember || oldStatus == telego.MemberStatusAdministrator) &&
			(newStatus == telego.MemberStatusLeft || newStatus == telego.MemberStatusBanned)

		switch {
		case isJoining:
			if _, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: chatID},
				Text:   MsgWelcome,
			}); err != nil {
				log.Printf("SendMessage welcome error (chat=%d): %v", chatID, err)
			}
			if err := bot.service.StartChallenge(ctx, chatID, 3, 180); err != nil {
				log.Printf("StartChallenge error (chat=%d): %v", chatID, err)
			}

		case isLeaving:
			if err := bot.service.DeactivateChallengeForChat(ctx, chatID); err != nil {
				log.Printf("DeactivateChallengeForChat error (chat=%d): %v", chatID, err)
			}
		}

		return nil
	})

	// /restart → інформація як перезапустити
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		_, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   MsgRestartInfo,
		})
		if err != nil {
			log.Printf("SendMessage restart info error (chat=%d): %v", message.Chat.ID, err)
		}
		return nil
	}, th.CommandEqual("restart"))

	// Щотижневий шедулер: чекає до кінця поточного тижня, запускає перевірку і нотифікує чат
	go bot.runWeeklyScheduler(ctx)

	bh.Start()
	return nil
}

// runWeeklyScheduler чекає до спливу поточного тижня челенджу,
// запускає WeeklyCheck, відправляє результат в чат і повторює цикл.
func (bot *Bot) runWeeklyScheduler(ctx context.Context) {
	for {
		chatID, nextCheck, err := bot.service.ActiveChallengeInfo(ctx)
		if err != nil {
			log.Printf("scheduler: no active challenge, retry in 1h: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Hour):
				continue
			}
		}

		log.Printf("scheduler: next weekly check at %s", nextCheck.Format(time.RFC3339))

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(nextCheck)):
		}

		failed, err := bot.service.WeeklyCheck(ctx)
		if err != nil {
			log.Printf("WeeklyCheck error: %v", err)
			continue
		}

		bot.notifyWeeklyResult(ctx, chatID, failed)
	}
}

func (bot *Bot) notifyWeeklyResult(ctx context.Context, chatID int64, failed []service.UserInfo) {
	var text string
	if len(failed) == 0 {
		text = MsgWeeklyAllGood
	} else {
		mentions := ""
		for _, u := range failed {
			mentions += fmt.Sprintf("@%s ", u.Username)
		}
		text = fmt.Sprintf(MsgWeeklyFailed, mentions)
	}

	if _, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   text,
	}); err != nil {
		log.Printf("notifyWeeklyResult error: %v", err)
	}
}
