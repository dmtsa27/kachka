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
func (c *challengeService) ActiveChallenges(ctx context.Context) ([]domain.Challenge, error) {
	challenges, err := c.challenge.GetAllActiveChallenges(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all active challenges: %w", err)
	}
	return challenges, nil
}

func (c *challengeService) WeeklyCheck(ctx context.Context, challenge domain.Challenge) ([]UserInfo, error) {
	elapsed := time.Since(challenge.StartedAt)
	weekNumber := int(elapsed / (7 * 24 * time.Hour))
	weekStart := challenge.StartedAt.Add(time.Duration(weekNumber) * 7 * 24 * time.Hour)

	userCounts, err := c.workouts.GetWorkoutCounts(ctx, weekStart)
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
