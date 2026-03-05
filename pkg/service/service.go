package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/storage"
)

type Repository interface {
	// users
	CreateUser(ctx context.Context, user storage.User) error
	ReadUser(ctx context.Context, telegramID int64) (storage.User, error)
	GetAllActiveUsers(ctx context.Context) ([]storage.User, error)
	DeactivateUser(ctx context.Context, telegramID int64) error

	// sessions
	HasTrainedToday(ctx context.Context, userID int64) (bool, error)
	StartSession(ctx context.Context, userID int64, chatID int64, messageID int) error
	GetSession(ctx context.Context, userID int64) (storage.Session, error)
	AddLatestSession(ctx context.Context, userID int64) error
	DeleteSessionToday(ctx context.Context, chatID int64, messageID int) error

	// workouts
	HasWorkoutToday(ctx context.Context, userID int64) (bool, error)
	CreateWorkout(ctx context.Context, workout storage.Workout) error
	WeeklyWorkouts(ctx context.Context, userID int64) (int, error)

	// challenge
	GetActiveChallenge(ctx context.Context) (storage.Challenge, error)
}

const (
	MinCircleDuration = 30 // seconds
	SessionGapMinutes = 20 // minutes between circles to count as workout
)

type Service struct {
	storage Repository
}

func New(s Repository) *Service {
	return &Service{storage: s}
}

// Called when user reacts with target emoji
func (s *Service) RegisterUser(ctx context.Context, telegramID int64, username string) error {
	return s.storage.CreateUser(ctx, storage.User{
		TelegramID: telegramID,
		Username:   username,
		IsActive:   true,
	})
}

// Called on every circle video message
func (s *Service) HandleCircle(ctx context.Context, userID int64, duration int, chatID int64, messageID int) error {
	// 1. Circle must be longer than 50 seconds
	if duration < MinCircleDuration {
		return nil
	}

	// 2. User must be registered (reacted with emoji)
	user, err := s.storage.ReadUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if !user.IsActive {
		return nil
	}

	// 3. Already has a completed workout today? Skip
	hasWorkout, err := s.storage.HasWorkoutToday(ctx, userID)
	if err != nil {
		return err
	}
	if hasWorkout {
		return nil
	}

	// 4. First circle today? Start a session
	hasSession, err := s.storage.HasTrainedToday(ctx, userID)
	if err != nil {
		return err
	}
	if !hasSession {
		return s.storage.StartSession(ctx, userID, chatID, messageID)
	}

	// 5. Session exists — check time gap between first and last circle
	session, err := s.storage.GetSession(ctx, userID)
	if err != nil {
		return err
	}

	if session.Last_video_at.Sub(session.Started_at) >= SessionGapMinutes*time.Minute {
		// 20+ min gap between first and last circle — completed workout!
		err = s.storage.CreateWorkout(ctx, storage.Workout{
			UserID:      userID,
			WorkoutDate: time.Now(),
		})
		if err != nil {
			return err
		}
	}

	// Always update last_video_at
	return s.storage.AddLatestSession(ctx, userID)
}

func (s *Service) CancelSession(ctx context.Context, chatID int64, messageID int) error {
	return s.storage.DeleteSessionToday(ctx, chatID, messageID)
}

// Called every Monday — returns users who failed the weekly goal
func (s *Service) WeeklyCheck(ctx context.Context) ([]storage.User, error) {
	challenge, err := s.storage.GetActiveChallenge(ctx)
	if err != nil {
		return nil, fmt.Errorf("no active challenge: %w", err)
	}

	users, err := s.storage.GetAllActiveUsers(ctx)
	if err != nil {
		return nil, err
	}

	var failed []storage.User
	for _, user := range users {
		count, err := s.storage.WeeklyWorkouts(ctx, user.TelegramID)
		if err != nil {
			return nil, err
		}
		if count < challenge.DaysPerWeek {
			if err := s.storage.DeactivateUser(ctx, user.TelegramID); err != nil {
				return nil, err
			}
			failed = append(failed, user)
		}
	}

	return failed, nil
}
