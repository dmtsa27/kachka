package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
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
	ProcessReactionUpdate(ctx context.Context, chatID int64, messageID int, userID int64, username string, emojis []string) (challengeStarted bool, workoutCancelled bool, err error)
	TryStartChallengeIfReady(ctx context.Context, chatID int64) (bool, error)
	DeactivateChallengeForChat(ctx context.Context, chatID int64) error
	WeeklyCheck(ctx context.Context, challenge domain.Challenge) ([]service.UserInfo, error)
	ActiveChallenges(ctx context.Context) ([]domain.Challenge, error)
	UpdateChallengeConfig(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error
	GetChallengeConfig(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error)
	GetRules() service.Rules
}

type Bot struct {
	client  *telego.Bot
	service BotService
	outbox  chan outboxMsg
}

type outboxMsg struct {
	chatID int64
	text   string
}

func (bot *Bot) SetService(svc BotService) {
	bot.service = svc
}

func (bot *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	select {
	case bot.outbox <- outboxMsg{chatID: chatID, text: text}:
		return nil
	default:
		log.Printf("outbox full, dropping message to %d: %s", chatID, text)
		return nil
	}
}

func New(token string, svc BotService) (*Bot, error) {
	bot, err := telego.NewBot(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		client:  bot,
		service: svc,
		outbox:  make(chan outboxMsg, 100),
	}, nil
}

func (bot *Bot) messageWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-bot.outbox:
			_, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: msg.chatID},
				Text:   msg.text,
			})
			if err != nil {
				log.Printf("failed to send async message to %d: %v", msg.chatID, err)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (bot *Bot) Start(ctx context.Context) error {
	go bot.messageWorker(ctx)

	if err := bot.client.DeleteWebhook(ctx, &telego.DeleteWebhookParams{}); err != nil {
		log.Printf("DeleteWebhook warning: %v", err)
	}

	updates, err := bot.client.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		AllowedUpdates: []string{
			"message",
			"message_reaction",
			"my_chat_member",
			"chat_member",
			"callback_query",
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

	// /restart → інформація як перезапустити
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		_ = bot.SendMessage(ctx, message.Chat.ID, MsgRestartInfo)
		return nil
	}, th.CommandEqual("restart"))

	// /days [число]
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		config, err := bot.service.GetChallengeConfig(ctx, message.Chat.ID)
		if err != nil {
			return err
		}
		if config.IsStarted || config.WelcomeMessageID != 0 {
			_ = bot.SendMessage(ctx, message.Chat.ID, MsgConfigLocked)
			return nil
		}

		fields := strings.Fields(message.Text)
		if len(fields) < 2 {
			_ = bot.SendMessage(ctx, message.Chat.ID, "❌ Вкажіть кількість днів, наприклад: /days 3")
			return nil
		}
		var val int
		if _, err := fmt.Sscanf(fields[1], "%d", &val); err != nil || val < 1 || val > 7 {
			_ = bot.SendMessage(ctx, message.Chat.ID, "❌ Вкажіть число від 1 до 7")
			return nil
		}
		config.DaysPerWeek = val
		if err := bot.service.UpdateChallengeConfig(ctx, message.Chat.ID, config.DaysPerWeek, config.DurationDays, config.Price); err != nil {
			return err
		}
		_ = bot.SendMessage(ctx, message.Chat.ID, fmt.Sprintf("✅ Встановлено %d тренування на тиждень. Перевірте /settings", val))
		return nil
	}, th.CommandEqual("days"))

	// /duration [число]
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		config, err := bot.service.GetChallengeConfig(ctx, message.Chat.ID)
		if err != nil {
			return err
		}
		if config.IsStarted || config.WelcomeMessageID != 0 {
			_ = bot.SendMessage(ctx, message.Chat.ID, MsgConfigLocked)
			return nil
		}

		fields := strings.Fields(message.Text)
		if len(fields) < 2 {
			_ = bot.SendMessage(ctx, message.Chat.ID, "❌ Вкажіть тривалість, наприклад: /duration 180")
			return nil
		}
		var val int
		if _, err := fmt.Sscanf(fields[1], "%d", &val); err != nil || val < 1 {
			_ = bot.SendMessage(ctx, message.Chat.ID, "❌ Вкажіть додатнє число")
			return nil
		}
		config.DurationDays = val
		if err := bot.service.UpdateChallengeConfig(ctx, message.Chat.ID, config.DaysPerWeek, config.DurationDays, config.Price); err != nil {
			return err
		}
		_ = bot.SendMessage(ctx, message.Chat.ID, fmt.Sprintf("✅ Встановлено тривалість %d днів. Перевірте /settings", val))
		return nil
	}, th.CommandEqual("duration"))

	// /penalty [число]
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		config, err := bot.service.GetChallengeConfig(ctx, message.Chat.ID)
		if err != nil {
			return err
		}
		if config.IsStarted || config.WelcomeMessageID != 0 {
			_ = bot.SendMessage(ctx, message.Chat.ID, MsgConfigLocked)
			return nil
		}

		fields := strings.Fields(message.Text)
		if len(fields) < 2 {
			_ = bot.SendMessage(ctx, message.Chat.ID, "❌ Вкажіть суму штрафу, наприклад: /penalty 500")
			return nil
		}
		var val int
		if _, err := fmt.Sscanf(fields[1], "%d", &val); err != nil || val < 0 {
			_ = bot.SendMessage(ctx, message.Chat.ID, "❌ Вкажіть число >= 0")
			return nil
		}
		config.Price = val
		if err := bot.service.UpdateChallengeConfig(ctx, message.Chat.ID, config.DaysPerWeek, config.DurationDays, config.Price); err != nil {
			return err
		}
		_ = bot.SendMessage(ctx, message.Chat.ID, fmt.Sprintf("✅ Встановлено штраф %d грн. Перевірте /settings", val))
		return nil
	}, th.CommandEqual("penalty"))

	// /settings → налаштування челенджу
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		chatID := message.Chat.ID
		config, err := bot.service.GetChallengeConfig(ctx, chatID)
		if err != nil {
			log.Printf("GetChallengeConfig error: %v", err)
			_ = bot.SendMessage(ctx, chatID, "❌ Налаштування не знайдені. Спробуйте пізніше або перезапустіть бота.")
			return nil
		}

		var text string
		var keyboard *telego.InlineKeyboardMarkup

		if config.IsStarted || config.WelcomeMessageID != 0 {
			text = fmt.Sprintf(MsgSettingsLocked, config.DaysPerWeek, config.DurationDays, config.Price)
		} else {
			text = fmt.Sprintf(MsgSettings, config.DaysPerWeek, config.DurationDays, config.Price)
			keyboard = &telego.InlineKeyboardMarkup{
				InlineKeyboard: [][]telego.InlineKeyboardButton{
					{
						{Text: "✅ Підтвердити налаштування", CallbackData: "confirm_config"},
					},
				},
			}
		}

		params := &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   text,
		}
		if keyboard != nil {
			params.ReplyMarkup = keyboard
		}

		_, err = bot.client.SendMessage(ctx, params)
		return err
	}, th.CommandEqual("settings"))

	// Callback queries for settings
	bh.HandleCallbackQuery(func(ctx *th.Context, query telego.CallbackQuery) error {
		chatID := query.Message.GetChat().ID
		config, err := bot.service.GetChallengeConfig(ctx, chatID)
		if err != nil {
			return err
		}

		if query.Data == "confirm_config" {
			if config.IsStarted || config.WelcomeMessageID != 0 {
				return bot.client.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
					CallbackQueryID: query.ID,
					Text:            MsgConfigLocked,
					ShowAlert:       true,
				})
			}
			// Verify bot is admin before proceeding
			member, err := bot.client.GetChatMember(ctx, &telego.GetChatMemberParams{
				ChatID: telego.ChatID{ID: chatID},
				UserID: bot.client.ID(),
			})
			isAdmin := false
			if err == nil && member != nil {
				status := member.MemberStatus()
				isAdmin = status == telego.MemberStatusAdministrator || status == telego.MemberStatusCreator
			}

			if !isAdmin {
				return bot.client.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
					CallbackQueryID: query.ID,
					Text:            "❌ Будь ласка, зробіть бота адміністратором, щоб він міг бачити реакції!",
					ShowAlert:       true,
				})
			}

			// Send official rules with variables
			rules := bot.service.GetRules()
			sessionGapMin := int(rules.SessionGap.Minutes())
			minCircleSec := rules.MinCircleDurationSeconds

			rulesText := fmt.Sprintf(MsgWelcome,
				config.DurationDays, config.Price, config.DurationDays,
				config.DaysPerWeek, config.DaysPerWeek,
				sessionGapMin, minCircleSec,
				config.DaysPerWeek, config.DaysPerWeek, config.Price,
				config.DurationDays,
			)

			welcomeMsg, err := bot.client.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: chatID},
				Text:   rulesText,
			})
			if err != nil {
				log.Printf("confirm_config SendMessage error: %v", err)
				return bot.client.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
					CallbackQueryID: query.ID,
					Text:            "❌ Помилка при надсиланні правил. Перевірте дозволи бота.",
					ShowAlert:       true,
				})
			}

			expectedReactions := 1
			memberCount, err := bot.client.GetChatMemberCount(ctx, &telego.GetChatMemberCountParams{
				ChatID: telego.ChatID{ID: chatID},
			})
			if err == nil && memberCount != nil {
				expectedReactions = *memberCount - 1
				if expectedReactions < 1 {
					expectedReactions = 1
				}
			}

			// Update bootstrap with welcome message ID, set as admin, and freeze roster
			if err := bot.service.InitChallengeBootstrap(ctx, chatID, welcomeMsg.MessageID, true, expectedReactions); err != nil {
				log.Printf("confirm_config InitChallengeBootstrap error: %v", err)
				return bot.client.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
					CallbackQueryID: query.ID,
					Text:            "❌ Помилка збереження налаштувань.",
					ShowAlert:       true,
				})
			}

			// Remove the confirm button to prevent double-clicks
			_, _ = bot.client.EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
				ChatID:    telego.ChatID{ID: chatID},
				MessageID: query.Message.GetMessageID(),
			})

			return bot.client.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "✅ Налаштування підтверджено! Чекаємо на ❤️",
			})
		}

		return bot.client.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
	})

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

		started, cancelled, err := bot.service.ProcessReactionUpdate(ctx, reaction.Chat.ID, reaction.MessageID, userID, username, emojis)
		if err != nil {
			log.Printf("ProcessReactionUpdate error (chat=%d user=%d msg=%d): %v", reaction.Chat.ID, userID, reaction.MessageID, err)
			return nil
		}

		if started {
			_ = bot.SendMessage(ctx, reaction.Chat.ID, MsgChallengeStarted)
		}

		if cancelled {
			_ = bot.SendMessage(ctx, reaction.Chat.ID, MsgWorkoutCancelled)
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
			_, _ = bot.client.SendSticker(ctx, &telego.SendStickerParams{
				ChatID:  telego.ChatID{ID: chatID},
				Sticker: telego.InputFile{FileID: DuckStickerFileID},
			})
			_, _ = bot.client.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: chatID},
				Text:   MsgOnboarding,
			})

			if err := bot.service.InitChallengeBootstrap(ctx, chatID, 0, isAdminNow, 1); err != nil {
				log.Printf("InitChallengeBootstrap error (chat=%d): %v", chatID, err)
			}

		case !isLeaving && isMemberActive(newStatus):
			if err := bot.service.SetBotAdminStatus(ctx, chatID, isAdminNow); err != nil {
				log.Printf("SetBotAdminStatus error (chat=%d): %v", chatID, err)
			}

			if started, err := bot.service.TryStartChallengeIfReady(ctx, chatID); err != nil {
				log.Printf("TryStartChallengeIfReady error (chat=%d): %v", chatID, err)
			} else if started {
				_ = bot.SendMessage(ctx, chatID, MsgChallengeStarted)
			}

		case isLeaving:
			if err := bot.service.DeactivateChallengeForChat(ctx, chatID); err != nil {
				log.Printf("DeactivateChallengeForChat error (chat=%d): %v", chatID, err)
			}
		}

		return nil
	})

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

// runWeeklyScheduler iterates over all active challenges, determines the next check time,
// and runs WeeklyCheck for each when due.
func (bot *Bot) runWeeklyScheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		challenges, err := bot.service.ActiveChallenges(ctx)
		if err != nil {
			log.Printf("scheduler: failed to fetch active challenges: %v", err)
		} else {
			for _, ch := range challenges {
				elapsed := time.Since(ch.StartedAt)
				weekNumber := int(elapsed / (7 * 24 * time.Hour))
				nextCheck := ch.StartedAt.Add(time.Duration(weekNumber+1) * 7 * 24 * time.Hour)

				if time.Now().After(nextCheck) {
					log.Printf("scheduler: running weekly check for chat %d", ch.ChatID)
					failed, err := bot.service.WeeklyCheck(ctx, ch)
					if err != nil {
						log.Printf("WeeklyCheck error (chat=%d): %v", ch.ChatID, err)
						continue
					}
					bot.notifyWeeklyResult(ctx, ch.ChatID, failed)
				}
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
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

	_ = bot.SendMessage(ctx, chatID, text)
}
