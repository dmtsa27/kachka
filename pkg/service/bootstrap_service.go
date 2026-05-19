package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

type bootstrapService struct {
	users      CircleBootstrapUseCase
	bootstrap  BootstrapRepository
	moderation ModerationRepository
	challenge  ChallengeRepository
}

func (b *bootstrapService) InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
	return b.bootstrap.InitChallengeBootstrap(ctx, chatID, welcomeMessageID, isBotAdmin, expectedReactions)
}

func (b *bootstrapService) UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error {
	return b.bootstrap.UpsertChatMember(ctx, chatID, userID, isBot, isActive)
}

func (b *bootstrapService) SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error {
	return b.bootstrap.SetBotAdminStatus(ctx, chatID, isBotAdmin)
}

func (b *bootstrapService) ProcessReactionUpdate(ctx context.Context, chatID int64, messageID int, userID int64, username string, emojis []string) (challengeStarted bool, workoutCancelled bool, err error) {
	if err := b.users.RegisterUser(ctx, userID, username); err != nil {
		return false, false, err
	}

	if err := b.UpsertChatMember(ctx, chatID, userID, false, true); err != nil {
		return false, false, err
	}

	if err := b.bootstrap.SetUserMessageReactions(ctx, chatID, messageID, userID, emojis); err != nil {
		return false, false, err
	}

	if containsEmoji(emojis, "👎") {
		cancelled, err := b.moderation.CancelCountedByMessage(ctx, chatID, messageID)
		if err != nil {
			return false, false, err
		}
		workoutCancelled = cancelled
	}

	bootstrap, err := b.bootstrap.GetChallengeBootstrap(ctx, chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, workoutCancelled, nil
		}
		return false, workoutCancelled, err
	}

	if bootstrap.WelcomeMessageID != messageID {
		return false, workoutCancelled, nil
	}

	started, err := b.TryStartChallengeIfReady(ctx, chatID)
	if err != nil {
		return false, workoutCancelled, err
	}

	return started, workoutCancelled, nil
}

func (b *bootstrapService) TryStartChallengeIfReady(ctx context.Context, chatID int64) (bool, error) {
	bootstrap, err := b.bootstrap.GetChallengeBootstrap(ctx, chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	if bootstrap.IsStarted || !bootstrap.IsBotAdmin {
		return false, nil
	}

	if bootstrap.ExpectedReactions <= 0 {
		return false, nil
	}

	heartMembers, err := b.bootstrap.CountWelcomeHeartReactions(ctx, chatID)
	if err != nil {
		return false, err
	}

	if heartMembers < bootstrap.ExpectedReactions {
		return false, nil
	}

	marked, err := b.bootstrap.MarkChallengeStarted(ctx, chatID)
	if err != nil {
		return false, err
	}
	if !marked {
		return false, nil
	}

	if err := b.challenge.CreateChallenge(ctx, domain.Challenge{
		ChatID:      chatID,
		DaysPerWeek: bootstrap.DaysPerWeek,
		Duration:    bootstrap.DurationDays,
		Price:       bootstrap.Price,
		IsActive:    true,
	}); err != nil {
		return false, err
	}

	return true, nil
}

func (b *bootstrapService) UpdateConfig(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
	return b.bootstrap.UpdateChallengeBootstrapConfig(ctx, chatID, daysPerWeek, durationDays, price)
}

func (b *bootstrapService) GetConfig(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
	return b.bootstrap.GetChallengeBootstrap(ctx, chatID)
}

func containsEmoji(emojis []string, target string) bool {
	for _, emoji := range emojis {
		if emoji == target {
			return true
		}
	}
	return false
}
