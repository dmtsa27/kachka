package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

type challengeService struct {
	challenge ChallengeRepository
	users     UserRepository
	workouts  WorkoutRepository
}

// StartChallenge deactivates current challenge and starts a new one.
func (c *challengeService) StartChallenge(ctx context.Context, chatID int64, daysPerWeek int, duration int) error {
	return c.challenge.CreateChallenge(ctx, domain.Challenge{
		ChatID:      chatID,
		DaysPerWeek: daysPerWeek,
		Duration:    duration,
		IsActive:    true,
	})
}

// DeactivateChallengeForChat clears the active challenge for a chat.
func (c *challengeService) DeactivateChallengeForChat(ctx context.Context, chatID int64) error {
	return c.challenge.DeactivateChallengeForChat(ctx, chatID)
}

// ActiveChallenges returns all active challenges.
func (c *challengeService) GetChallenge(ctx context.Context, challengeID int) (*domain.Challenge, error) {
	return c.challenge.GetChallenge(ctx, challengeID)
}

func (c *challengeService) UpdateChallenge(ctx context.Context, challenge domain.Challenge) error {
	return c.challenge.UpdateChallenge(ctx, challenge)
}

func (c *challengeService) DeleteChallenge(ctx context.Context, challengeID int) error {
	return c.challenge.DeleteChallenge(ctx, challengeID)
}

func (c *challengeService) ActiveChallenges(ctx context.Context) ([]domain.Challenge, error) {
	challenges, err := c.challenge.GetAllActiveChallenges(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all active challenges: %w", err)
	}
	return challenges, nil
}

func getWeekStart(t time.Time) time.Time {
	// Normalize to UTC midnight
	t = t.UTC()
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	// Find previous Monday
	offset := int(t.Weekday()) - 1
	if offset < 0 {
		offset = 6 // Sunday -> Monday is -6 days
	}
	return t.AddDate(0, 0, -offset)
}

func (c *challengeService) WeeklyCheck(ctx context.Context, challenge domain.Challenge) ([]UserInfo, error) {
	weekStart := getWeekStart(time.Now())
	// If the challenge started this week, use the previous week's start to avoid skipping check
	// OR better: check if it's already time to check based on challenge.StartedAt.
	// For now, let's just use the current Monday.

	userCounts, err := c.workouts.GetWorkoutCounts(ctx, challenge.ChatID, weekStart)
	if err != nil {
		return nil, fmt.Errorf("get workout counts: %w", err)
	}

	var failedIDs []int64
	var failed []UserInfo
	for _, uc := range userCounts {
		if uc.Count < challenge.DaysPerWeek {
			failedIDs = append(failedIDs, uc.TelegramID)
			failed = append(failed, uc.UserInfo)
		}
	}

	if len(failedIDs) > 0 {
		if err := c.users.BatchDeactivateUsers(ctx, failedIDs); err != nil {
			return nil, fmt.Errorf("batch deactivate: %w", err)
		}
	}

	return failed, nil
}

func (c *challengeService) MarkWeeklyCheckDone(ctx context.Context, challengeID int) error {
	return c.challenge.MarkWeeklyCheckDone(ctx, challengeID)
}

func (c *challengeService) MarkDailyStatsDone(ctx context.Context, challengeID int) error {
	return c.challenge.MarkDailyStatsDone(ctx, challengeID)
}

func (c *challengeService) AddWorkoutDirect(ctx context.Context, chatID int64, username string, amount int) (int, error) {
	userID, err := c.users.GetUserIDByUsername(ctx, username)
	if err != nil {
		return 0, err
	}

	return c.workouts.AddWorkouts(ctx, userID, chatID, amount)
}

func (c *challengeService) GetStats(ctx context.Context, chatID int64) ([]domain.UserStats, error) {
	_, err := c.challenge.GetActiveChallengeByChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get active challenge: %w", err)
	}

	weekStart := getWeekStart(time.Now())
	return c.workouts.GetChatStats(ctx, chatID, weekStart)
}
