package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/storage"
)

// UserInfo — domain DTO, не залежить від інфраструктури.
type UserInfo struct {
	TelegramID int64
	Username   string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user storage.User) error
	ReadUser(ctx context.Context, telegramID int64) (*storage.User, error)
	GetAllActiveUsers(ctx context.Context) ([]storage.User, error)
	// BatchDeactivateUsers деактивує всіх переданих юзерів в одній транзакції.
	BatchDeactivateUsers(ctx context.Context, userIDs []int64) error
}

type SessionRepository interface {
	HasTrainedToday(ctx context.Context, userID int64) (bool, error)
	StartSession(ctx context.Context, userID int64, chatID int64, messageID int) error
	GetSession(ctx context.Context, userID int64) (*storage.Session, error)
	AddLatestSession(ctx context.Context, userID int64) error
	DeleteSessionToday(ctx context.Context, chatID int64, messageID int) error
}

type WorkoutRepository interface {
	HasWorkoutToday(ctx context.Context, userID int64) (bool, error)
	CreateWorkout(ctx context.Context, workout storage.Workout) error
	WeeklyWorkouts(ctx context.Context, userID int64, weekStart time.Time) (int, error)
}

type BootstrapRepository interface {
	InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error
	GetChallengeBootstrap(ctx context.Context, chatID int64) (*storage.ChallengeBootstrap, error)
	SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error
	UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error
	SetUserMessageReactions(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error
	CountWelcomeHeartReactions(ctx context.Context, chatID int64) (int, error)
	MarkChallengeStarted(ctx context.Context, chatID int64) (bool, error)
}

type ModerationRepository interface {
	CancelCountedByMessage(ctx context.Context, chatID int64, messageID int) (bool, error)
}

type ChallengeRepository interface {
	GetActiveChallenge(ctx context.Context) (*storage.Challenge, error)
	CreateChallenge(ctx context.Context, challenge storage.Challenge) error
	DeactivateChallengeForChat(ctx context.Context, chatID int64) error
}

type Repositories interface {
	UserRepository
	SessionRepository
	WorkoutRepository
	ChallengeRepository
	BootstrapRepository
	ModerationRepository
}

// Notifier відповідає за відправку повідомлень в чат.
// Імплементується в пакеті telegram.
type Notifier interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

const (
	MinCircleDuration = 30 // seconds
	SessionGapMinutes = 20 // minutes between circles to count as workout
)

type Service struct {
	users     UserRepository
	sessions  SessionRepository
	workouts  WorkoutRepository
	challenge ChallengeRepository
	bootstrap BootstrapRepository
	moderate  ModerationRepository
	notifier  Notifier
}

func New(r Repositories, n Notifier) *Service {
	return &Service{
		users:     r,
		sessions:  r,
		workouts:  r,
		challenge: r,
		bootstrap: r,
		moderate:  r,
		notifier:  n,
	}
}

func (s *Service) RegisterUser(ctx context.Context, telegramID int64, username string) error {
	return s.users.CreateUser(ctx, storage.User{
		TelegramID: telegramID,
		Username:   username,
		IsActive:   true,
	})
}

func (s *Service) HandleCircle(ctx context.Context, userID int64, duration int, chatID int64, messageID int) error {
	if duration < MinCircleDuration {
		return nil
	}

	active, err := s.isActiveUser(ctx, userID)
	if err != nil {
		return err
	}
	if !active {
		return nil
	}

	hasWorkout, err := s.workouts.HasWorkoutToday(ctx, userID)
	if err != nil {
		return err
	}
	if hasWorkout {
		return nil
	}

	return s.processSession(ctx, userID, chatID, messageID)
}

func (s *Service) isActiveUser(ctx context.Context, userID int64) (bool, error) {
	user, err := s.users.ReadUser(ctx, userID)
	if err != nil {
		// Незареєстрований юзер — нормальна ситуація, не помилка.
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("read user: %w", err)
	}
	return user.IsActive, nil
}

func (s *Service) processSession(ctx context.Context, userID int64, chatID int64, messageID int) error {
	hasSession, err := s.sessions.HasTrainedToday(ctx, userID)
	if err != nil {
		return err
	}
	if !hasSession {
		if err := s.sessions.StartSession(ctx, userID, chatID, messageID); err != nil {
			return err
		}
		if s.notifier != nil {
			_ = s.notifier.SendMessage(ctx, chatID, "🚀")
		}
		return nil
	}

	if err = s.sessions.AddLatestSession(ctx, userID); err != nil {
		return err
	}

	session, err := s.sessions.GetSession(ctx, userID)
	if err != nil {
		return err
	}

	if s.isWorkoutComplete(*session) {
		if err = s.workouts.CreateWorkout(ctx, storage.Workout{
			UserID:      userID,
			WorkoutDate: time.Now(),
			ChatID:      chatID,
			MessageID:   messageID,
		}); err != nil {
			return err
		}
		if s.notifier != nil {
			_ = s.notifier.SendMessage(ctx, chatID, "🎖")
		}
	}

	return nil
}

func (s *Service) isWorkoutComplete(session storage.Session) bool {
	return session.LastVideoAt.Sub(session.StartedAt) >= SessionGapMinutes*time.Minute
}

func (s *Service) CancelSession(ctx context.Context, chatID int64, messageID int) error {
	return s.sessions.DeleteSessionToday(ctx, chatID, messageID)
}

// StartChallenge деактивує поточний челендж і стартує новий.
func (s *Service) StartChallenge(ctx context.Context, chatID int64, daysPerWeek int, duration int) error {
	return s.challenge.CreateChallenge(ctx, storage.Challenge{
		ChatID:      chatID,
		DaysPerWeek: daysPerWeek,
		Duration:    duration,
		IsActive:    true,
	})
}

// DeactivateChallengeForChat очищає активний челендж чату.
// Викликається коли бота видаляють з групи або чат видаляється.
func (s *Service) DeactivateChallengeForChat(ctx context.Context, chatID int64) error {
	return s.challenge.DeactivateChallengeForChat(ctx, chatID)
}

// ActiveChallengeInfo повертає chatID активного челенджу і час, коли потрібно
// зробити наступну щотижневу перевірку.
func (s *Service) ActiveChallengeInfo(ctx context.Context) (chatID int64, nextCheck time.Time, err error) {
	challenge, err := s.challenge.GetActiveChallenge(ctx)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("no active challenge: %w", err)
	}

	elapsed := time.Since(challenge.StartedAt)
	weekNumber := int(elapsed / (7 * 24 * time.Hour))
	nextCheck = challenge.StartedAt.Add(time.Duration(weekNumber+1) * 7 * 24 * time.Hour)

	return challenge.ChatID, nextCheck, nil
}

func (s *Service) WeeklyCheck(ctx context.Context) ([]UserInfo, error) {
	challenge, err := s.challenge.GetActiveChallenge(ctx)
	if err != nil {
		return nil, fmt.Errorf("no active challenge: %w", err)
	}

	// Рахуємо номер поточного тижня відносно старту челенджу.
	// Тиждень 0: [started_at, started_at+7d), тиждень 1: [started_at+7d, started_at+14d) і т.д.
	elapsed := time.Since(challenge.StartedAt)
	weekNumber := int(elapsed / (7 * 24 * time.Hour))
	weekStart := challenge.StartedAt.Add(time.Duration(weekNumber) * 7 * 24 * time.Hour)

	users, err := s.users.GetAllActiveUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Спочатку збираємо всіх хто не виконав норму, потім деактивуємо в одній транзакції.
	var failedIDs []int64
	var failed []UserInfo
	for _, user := range users {
		count, err := s.workouts.WeeklyWorkouts(ctx, user.TelegramID, weekStart)
		if err != nil {
			return nil, err
		}
		if count < challenge.DaysPerWeek {
			failedIDs = append(failedIDs, user.TelegramID)
			failed = append(failed, UserInfo{TelegramID: user.TelegramID, Username: user.Username})
		}
	}

	if len(failedIDs) > 0 {
		if err := s.users.BatchDeactivateUsers(ctx, failedIDs); err != nil {
			return nil, fmt.Errorf("batch deactivate: %w", err)
		}
	}

	return failed, nil
}

// SetNotifier дозволяє встановити Notifier після створення Service,
// вирішуючи проблему циклічної залежності при ініціалізації.
func (s *Service) SetNotifier(n Notifier) {
	s.notifier = n
}

func (s *Service) InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
	return s.bootstrap.InitChallengeBootstrap(ctx, chatID, welcomeMessageID, isBotAdmin, expectedReactions)
}

func (s *Service) UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error {
	return s.bootstrap.UpsertChatMember(ctx, chatID, userID, isBot, isActive)
}

func (s *Service) SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error {
	return s.bootstrap.SetBotAdminStatus(ctx, chatID, isBotAdmin)
}

func (s *Service) ProcessReactionUpdate(ctx context.Context, chatID int64, messageID int, userID int64, username string, emojis []string, daysPerWeek int, duration int) (challengeStarted bool, workoutCancelled bool, err error) {
	if err := s.RegisterUser(ctx, userID, username); err != nil {
		return false, false, err
	}

	if err := s.UpsertChatMember(ctx, chatID, userID, false, true); err != nil {
		return false, false, err
	}

	if err := s.bootstrap.SetUserMessageReactions(ctx, chatID, messageID, userID, emojis); err != nil {
		return false, false, err
	}

	if containsEmoji(emojis, "👎") {
		cancelled, err := s.moderate.CancelCountedByMessage(ctx, chatID, messageID)
		if err != nil {
			return false, false, err
		}
		workoutCancelled = cancelled
	}

	bootstrap, err := s.bootstrap.GetChallengeBootstrap(ctx, chatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, workoutCancelled, nil
		}
		return false, workoutCancelled, err
	}

	if bootstrap.WelcomeMessageID != messageID {
		return false, workoutCancelled, nil
	}

	started, err := s.TryStartChallengeIfReady(ctx, chatID, daysPerWeek, duration)
	if err != nil {
		return false, workoutCancelled, err
	}

	return started, workoutCancelled, nil
}

func (s *Service) TryStartChallengeIfReady(ctx context.Context, chatID int64, daysPerWeek int, duration int) (bool, error) {
	bootstrap, err := s.bootstrap.GetChallengeBootstrap(ctx, chatID)
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

	heartMembers, err := s.bootstrap.CountWelcomeHeartReactions(ctx, chatID)
	if err != nil {
		return false, err
	}

	if heartMembers < bootstrap.ExpectedReactions {
		return false, nil
	}

	marked, err := s.bootstrap.MarkChallengeStarted(ctx, chatID)
	if err != nil {
		return false, err
	}
	if !marked {
		return false, nil
	}

	if err := s.challenge.CreateChallenge(ctx, storage.Challenge{
		ChatID:      chatID,
		DaysPerWeek: daysPerWeek,
		Duration:    duration,
		IsActive:    true,
	}); err != nil {
		return false, err
	}

	return true, nil
}

func containsEmoji(emojis []string, target string) bool {
	for _, emoji := range emojis {
		if emoji == target {
			return true
		}
	}
	return false
}
