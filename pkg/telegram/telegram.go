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
	UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error
	HandleCircle(ctx context.Context, userID int64, duration int, chatID int64, messageID int) error
	InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error
	SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error
	ProcessReactionUpdate(ctx context.Context, chatID int64, messageID int, userID int64, username string, emojis []string, daysPerWeek int, duration int) (challengeStarted bool, workoutCancelled bool, err error)
	TryStartChallengeIfReady(ctx context.Context, chatID int64, daysPerWeek int, duration int) (bool, error)
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
	if err := bot.client.DeleteWebhook(ctx, &telego.DeleteWebhookParams{}); err != nil {
		log.Printf("DeleteWebhook warning: %v", err)
	}

	updates, err := bot.client.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		AllowedUpdates: []string{
			"message",
			"message_reaction",
			"my_chat_member",
			"chat_member",
		},
	})
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
		if message.From != nil {
			if err := bot.service.RegisterUser(ctx, message.From.ID, message.From.Username); err != nil {
				log.Printf("RegisterUser error (user=%d): %v", message.From.ID, err)
			}
			if err := bot.service.UpsertChatMember(ctx, message.Chat.ID, message.From.ID, message.From.IsBot, true); err != nil {
				log.Printf("UpsertChatMember error (chat=%d user=%d): %v", message.Chat.ID, message.From.ID, err)
			}
		}

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

	// Reactions: register/update user, sync reactions, start challenge on all ❤️, cancel counted workout on 👎.
	bh.HandleMessageReaction(func(ctx *th.Context, reaction telego.MessageReactionUpdated) error {
		if reaction.User == nil {
			return nil
		}

		userID := reaction.User.ID
		username := reaction.User.Username
		emojis := extractEmojiReactions(reaction.NewReaction)

		started, cancelled, err := bot.service.ProcessReactionUpdate(ctx, reaction.Chat.ID, reaction.MessageID, userID, username, emojis, 3, 180)
		if err != nil {
			log.Printf("ProcessReactionUpdate error (chat=%d user=%d msg=%d): %v", reaction.Chat.ID, userID, reaction.MessageID, err)
			return nil
		}

		if started {
			if _, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: reaction.Chat.ID},
				Text:   MsgChallengeStarted,
			}); err != nil {
				log.Printf("SendMessage challenge started error (chat=%d): %v", reaction.Chat.ID, err)
			}
		}

		if cancelled {
			if _, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: reaction.Chat.ID},
				Text:   MsgWorkoutCancelled,
			}); err != nil {
				log.Printf("SendMessage workout cancelled error (chat=%d): %v", reaction.Chat.ID, err)
			}
		}
		return nil
	})

	// Track chat members to freeze roster at welcome moment.
	bh.HandleChatMemberUpdated(func(ctx *th.Context, update telego.ChatMemberUpdated) error {
		member := update.NewChatMember.MemberUser()
		active := isMemberActive(update.NewChatMember.MemberStatus())

		if err := bot.service.UpsertChatMember(ctx, update.Chat.ID, member.ID, member.IsBot, active); err != nil {
			log.Printf("UpsertChatMember update error (chat=%d user=%d): %v", update.Chat.ID, member.ID, err)
		}
		return nil
	})

	// Bot доданий в групу → привітальне повідомлення + bootstrap челенджу.
	// Бот кікнутий/чат видалений → деактивуємо челендж.
	bh.HandleMyChatMemberUpdated(func(ctx *th.Context, update telego.ChatMemberUpdated) error {
		oldStatus := update.OldChatMember.MemberStatus()
		newStatus := update.NewChatMember.MemberStatus()
		chatID := update.Chat.ID

		isJoining := isMemberInactive(oldStatus) && isMemberActive(newStatus)

		isLeaving := isMemberActive(oldStatus) && isMemberInactive(newStatus)

		isAdminNow := newStatus == telego.MemberStatusAdministrator || newStatus == telego.MemberStatusCreator

		switch {
		case isJoining:
			welcomeMsg, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: chatID},
				Text:   MsgWelcome,
			})
			if err != nil {
				log.Printf("SendMessage welcome error (chat=%d): %v", chatID, err)
				return nil
			}

			expectedReactions := 1
			memberCount, err := bot.client.GetChatMemberCount(ctx, &telego.GetChatMemberCountParams{
				ChatID: telego.ChatID{ID: chatID},
			})
			if err != nil {
				log.Printf("GetChatMemberCount warning (chat=%d): %v", chatID, err)
			} else if memberCount != nil {
				expectedReactions = *memberCount - 1
				if expectedReactions < 1 {
					expectedReactions = 1
				}
			}

			if err := bot.service.InitChallengeBootstrap(ctx, chatID, welcomeMsg.MessageID, isAdminNow, expectedReactions); err != nil {
				log.Printf("InitChallengeBootstrap error (chat=%d): %v", chatID, err)
			}

			if started, err := bot.service.TryStartChallengeIfReady(ctx, chatID, 3, 180); err != nil {
				log.Printf("TryStartChallengeIfReady error (chat=%d): %v", chatID, err)
			} else if started {
				if _, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
					ChatID: telego.ChatID{ID: chatID},
					Text:   MsgChallengeStarted,
				}); err != nil {
					log.Printf("SendMessage challenge started error (chat=%d): %v", chatID, err)
				}
			}

		case !isLeaving && isMemberActive(newStatus):
			if err := bot.service.SetBotAdminStatus(ctx, chatID, isAdminNow); err != nil {
				log.Printf("SetBotAdminStatus error (chat=%d): %v", chatID, err)
			}

			if started, err := bot.service.TryStartChallengeIfReady(ctx, chatID, 3, 180); err != nil {
				log.Printf("TryStartChallengeIfReady error (chat=%d): %v", chatID, err)
			} else if started {
				if _, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
					ChatID: telego.ChatID{ID: chatID},
					Text:   MsgChallengeStarted,
				}); err != nil {
					log.Printf("SendMessage challenge started error (chat=%d): %v", chatID, err)
				}
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

func isMemberActive(status string) bool {
	return status == telego.MemberStatusMember ||
		status == telego.MemberStatusAdministrator ||
		status == telego.MemberStatusRestricted ||
		status == telego.MemberStatusCreator
}

func isMemberInactive(status string) bool {
	return status == telego.MemberStatusLeft || status == telego.MemberStatusBanned
}

func extractEmojiReactions(reactions []telego.ReactionType) []string {
	emojis := make([]string, 0, len(reactions))
	for _, reaction := range reactions {
		emoji, ok := reaction.(*telego.ReactionTypeEmoji)
		if !ok {
			continue
		}
		emojis = append(emojis, emoji.Emoji)
	}
	return emojis
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
