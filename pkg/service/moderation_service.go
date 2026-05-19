package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

const (
	DisputeWindow = 10 * time.Minute
	VoteWindow    = 5 * time.Minute
)

type moderationService struct {
	workouts WorkoutRepository
	sessions SessionRepository
	votes    VoteRepository
	users    UserUseCase
	notifier Notifier
}

func (m *moderationService) DisputeWorkout(ctx context.Context, chatID int64, messageID int, disputerID int64) (int64, bool, error) {
	// 1. Try to find and cancel a completed workout
	workout, err := m.workouts.GetWorkoutByMessage(ctx, chatID, messageID)
	if err == nil {
		if workout.UserID == disputerID {
			return 0, false, errors.New("не можна скасовувати власне тренування")
		}

		if time.Since(workout.WorkoutDate) > DisputeWindow {
			return 0, false, errors.New("час для скасування вичерпано (10 хв)")
		}

		targetUserID, err := m.workouts.CancelWorkout(ctx, chatID, messageID, disputerID)
		if err != nil {
			return 0, false, fmt.Errorf("помилка бази даних: %w", err)
		}
		return targetUserID, false, nil
	}

	// 2. If no workout, try to find and cancel a session (rocket)
	session, err := m.getSessionByMessageID(ctx, chatID, messageID)
	if err == nil {
		if session.UserID == disputerID {
			return 0, true, errors.New("не можна скасовувати власний старт")
		}
		if time.Since(session.StartedAt) > DisputeWindow {
			return 0, true, errors.New("час для скасування старту вичерпано")
		}

		err = m.sessions.DeleteSessionToday(ctx, chatID, messageID)
		if err != nil {
			return 0, true, fmt.Errorf("помилка видалення сесії: %w", err)
		}
		return session.UserID, true, nil
	}

	return 0, false, errors.New("тренування або старт не знайдено")
}

func (m *moderationService) getSessionByMessageID(ctx context.Context, chatID int64, messageID int) (*domain.Session, error) {
	return m.sessions.GetSessionByMessage(ctx, chatID, messageID)
}

func (m *moderationService) ReinstateWorkout(ctx context.Context, chatID int64, messageID int, reinstaterID int64) error {
	workout, err := m.workouts.GetWorkoutByMessage(ctx, chatID, messageID)
	if err != nil {
		return errors.New("тренування не знайдено")
	}

	if workout.CancelledBy == nil || *workout.CancelledBy != reinstaterID {
		return errors.New("лише той, хто скасував, може повернути тренування")
	}

	if time.Since(workout.WorkoutDate) > DisputeWindow {
		return errors.New("час для повернення вичерпано")
	}

	return m.workouts.ReinstateWorkout(ctx, chatID, messageID, reinstaterID)
}

func (m *moderationService) InitiateSubtract(ctx context.Context, chatID int64, initiatorID int64, targetUsername string, amount int, pollID string) error {
	targetUserID, err := m.users.GetUserIDByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}

	if targetUserID == initiatorID {
		return errors.New("не можна голосувати проти себе")
	}

	vote := domain.Vote{
		ChatID:       chatID,
		TargetUserID: targetUserID,
		InitiatorID:  initiatorID,
		PollID:       pollID,
		Amount:       amount,
		Type:         "subtract",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(VoteWindow),
	}

	return m.votes.CreateVote(ctx, vote)
}

func (m *moderationService) InitiateAdd(ctx context.Context, chatID int64, initiatorID int64, targetUsername string, amount int, pollID string) error {
	targetUserID, err := m.users.GetUserIDByUsername(ctx, targetUsername)
	if err != nil {
		return err
	}

	vote := domain.Vote{
		ChatID:       chatID,
		TargetUserID: targetUserID,
		InitiatorID:  initiatorID,
		PollID:       pollID,
		Amount:       amount,
		Type:         "add",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(VoteWindow),
	}

	return m.votes.CreateVote(ctx, vote)
}

func (m *moderationService) HandlePollUpdate(ctx context.Context, pollID string, success bool) (int, error) {
	vote, err := m.votes.GetVoteByPollID(ctx, pollID)
	if err != nil {
		return 0, err
	}

	if vote.IsCompleted {
		return 0, nil
	}

	if err := m.votes.CompleteVote(ctx, pollID, success); err != nil {
		return 0, err
	}

	if success {
		if vote.Type == "add" {
			return m.workouts.AddWorkouts(ctx, vote.TargetUserID, vote.ChatID, vote.Amount)
		}
		return m.workouts.SubtractWorkouts(ctx, vote.TargetUserID, vote.ChatID, vote.Amount)
	}

	return 0, nil
}

func (m *moderationService) GetWorkoutByMessage(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
	return m.workouts.GetWorkoutByMessage(ctx, chatID, messageID)
}

func (m *moderationService) GetActiveChallengeVotersCount(ctx context.Context, chatID int64) (int, error) {
	return m.workouts.GetActiveChallengeVotersCount(ctx, chatID)
}
