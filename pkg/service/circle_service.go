package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

type circleService struct {
	users     CircleUserUseCase
	sessions  SessionRepository
	workouts  WorkoutRepository
	challenge ChallengeRepository
	notifier  Notifier
	rules     Rules
}

func (c *circleService) HandleCircle(ctx context.Context, userID int64, duration int, chatID int64, messageID int) error {
	if duration < c.rules.MinCircleDurationSeconds {
		return nil
	}

	isActiveChat, err := c.isActiveChallengeChat(ctx, chatID)
	if err != nil {
		return err
	}
	if !isActiveChat {
		return nil
	}

	active, err := c.users.IsActiveUser(ctx, userID)
	if err != nil {
		return err
	}
	if !active {
		return nil
	}

	hasWorkout, err := c.workouts.HasWorkoutToday(ctx, userID, chatID)
	if err != nil {
		return err
	}
	if hasWorkout {
		return nil
	}

	return c.processSession(ctx, userID, chatID, messageID)
}

func (c *circleService) isActiveChallengeChat(ctx context.Context, chatID int64) (bool, error) {
	isActive, err := c.challenge.HasActiveChallengeInChat(ctx, chatID)
	if err != nil {
		return false, fmt.Errorf("check active challenge in chat: %w", err)
	}

	return isActive, nil
}

func (c *circleService) processSession(ctx context.Context, userID int64, chatID int64, messageID int) error {
	hasSession, err := c.sessions.HasTrainedToday(ctx, userID, chatID)
	if err != nil {
		return err
	}
	if !hasSession {
		msgID := messageID
		if c.notifier != nil {
			id, err := c.notifier.SendMessage(ctx, chatID, "🚀")
			if err == nil {
				msgID = id
			}
		}
		if err := c.sessions.StartSession(ctx, userID, chatID, msgID); err != nil {
			return err
		}
		return nil
	}

	if err = c.sessions.AddLatestSession(ctx, userID, chatID); err != nil {
		return err
	}

	session, err := c.sessions.GetSession(ctx, userID, chatID)
	if err != nil {
		return err
	}

	if c.isWorkoutComplete(*session) {
		msgID := messageID
		if c.notifier != nil {
			id, err := c.notifier.SendMessage(ctx, chatID, "🎖")
			if err == nil {
				msgID = id
			}
		}

		if err = c.workouts.CreateWorkout(ctx, domain.Workout{
			UserID:      userID,
			WorkoutDate: time.Now(),
			ChatID:      chatID,
			MessageID:   msgID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *circleService) isWorkoutComplete(session domain.Session) bool {
	return session.LastVideoAt.Sub(session.StartedAt) >= c.rules.SessionGap
}

func (c *circleService) CancelSession(ctx context.Context, chatID int64, messageID int) {
	_ = c.sessions.DeleteSessionToday(ctx, chatID, messageID)
}
