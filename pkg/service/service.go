package service

import (
	"context"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

// UserInfo is a lightweight DTO for reporting results.
type UserInfo struct {
	TelegramID int64
	Username   string
}

// UserWorkouts combines user info with their workout count.
type UserWorkouts struct {
	UserInfo
	Count int
}

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) error
	ReadUser(ctx context.Context, telegramID int64) (*domain.User, error)
	GetAllActiveUsers(ctx context.Context) ([]domain.User, error)
	// BatchDeactivateUsers deactivates all passed users in one transaction.
	BatchDeactivateUsers(ctx context.Context, userIDs []int64) error
}

type SessionRepository interface {
	HasTrainedToday(ctx context.Context, userID int64, chatID int64) (bool, error)
	StartSession(ctx context.Context, userID int64, chatID int64, messageID int) error
	GetSession(ctx context.Context, userID int64, chatID int64) (*domain.Session, error)
	AddLatestSession(ctx context.Context, userID int64, chatID int64) error
	DeleteSessionToday(ctx context.Context, chatID int64, messageID int) error
}

type WorkoutRepository interface {
	HasWorkoutToday(ctx context.Context, userID int64, chatID int64) (bool, error)
	CreateWorkout(ctx context.Context, workout domain.Workout) error
	WeeklyWorkouts(ctx context.Context, userID int64, weekStart time.Time) (int, error)
	// GetWorkoutCounts returns workout counts for all active users since weekStart.
	GetWorkoutCounts(ctx context.Context, weekStart time.Time) ([]UserWorkouts, error)
}

type BootstrapRepository interface {
	InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error
	GetChallengeBootstrap(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error)
	SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error
	UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error
	SetUserMessageReactions(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error
	CountWelcomeHeartReactions(ctx context.Context, chatID int64) (int, error)
	MarkChallengeStarted(ctx context.Context, chatID int64) (bool, error)
	UpdateChallengeBootstrapConfig(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error
}

type ModerationRepository interface {
	CancelCountedByMessage(ctx context.Context, chatID int64, messageID int) (bool, error)
}

type ChallengeRepository interface {
	GetActiveChallenge(ctx context.Context) (*domain.Challenge, error)
	GetAllActiveChallenges(ctx context.Context) ([]domain.Challenge, error)
	HasActiveChallengeInChat(ctx context.Context, chatID int64) (bool, error)
	CreateChallenge(ctx context.Context, challenge domain.Challenge) error
	DeactivateChallengeForChat(ctx context.Context, chatID int64) error
}

// Notifier sends bot messages to the chat.
// Implemented by the telegram package.
type Notifier interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

const (
	MinCircleDuration = 30 // seconds
	SessionGapMinutes = 20 // minutes between circles to count as workout
)

// Rules controls thresholds without hardcoding them across handlers.
type Rules struct {
	MinCircleDurationSeconds int
	SessionGap               time.Duration
}

// DefaultRules preserves existing behavior.
func DefaultRules() Rules {
	return Rules{
		MinCircleDurationSeconds: MinCircleDuration,
		SessionGap:               time.Duration(SessionGapMinutes) * time.Minute,
	}
}

// Deps collects explicit dependencies for the service facade.
type Deps struct {
	Users      UserRepository
	Sessions   SessionRepository
	Workouts   WorkoutRepository
	Challenge  ChallengeRepository
	Bootstrap  BootstrapRepository
	Moderation ModerationRepository
	Notifier   Notifier
	Rules      Rules
}

// BootstrapUseCase defines operations for starting a challenge.
type BootstrapUseCase interface {
	InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error
	UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error
	SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error
	ProcessReactionUpdate(ctx context.Context, chatID int64, messageID int, userID int64, username string, emojis []string) (challengeStarted bool, workoutCancelled bool, err error)
	TryStartChallengeIfReady(ctx context.Context, chatID int64) (bool, error)
	UpdateConfig(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error
	GetConfig(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error)
}

// ChallengeUseCase defines operations for managing active challenges.
type ChallengeUseCase interface {
	StartChallenge(ctx context.Context, chatID int64, daysPerWeek int, duration int) error
	DeactivateChallengeForChat(ctx context.Context, chatID int64) error
	ActiveChallenges(ctx context.Context) ([]domain.Challenge, error)
	WeeklyCheck(ctx context.Context, challenge domain.Challenge) ([]UserInfo, error)
}

// CircleUserUseCase defines user operations needed by CircleUseCase.
type CircleUserUseCase interface {
	IsActiveUser(ctx context.Context, userID int64) (bool, error)
}

// CircleBootstrapUseCase defines bootstrap operations needed by BootstrapUseCase.
type CircleBootstrapUseCase interface {
	RegisterUser(ctx context.Context, telegramID int64, username string) error
}

// CircleUseCase defines operations for handling video messages.
type CircleUseCase interface {
	HandleCircle(ctx context.Context, userID int64, duration int, chatID int64, messageID int) error
	CancelSession(ctx context.Context, chatID int64, messageID int)
}

// UserUseCase defines operations for user management.
type UserUseCase interface {
	RegisterUser(ctx context.Context, telegramID int64, username string) error
}

// Service is a facade that delegates to focused use cases.
type Service struct {
	Users     UserUseCase
	Circle    CircleUseCase
	Challenge ChallengeUseCase
	Bootstrap BootstrapUseCase
	rules     Rules
}

func New(deps Deps) *Service {
	rules := normalizeRules(deps.Rules)

	userSvc := &userService{users: deps.Users}
	circleSvc := &circleService{
		users:     userSvc,
		sessions:  deps.Sessions,
		workouts:  deps.Workouts,
		challenge: deps.Challenge,
		notifier:  deps.Notifier,
		rules:     rules,
	}
	challengeSvc := &challengeService{
		challenge: deps.Challenge,
		users:     deps.Users,
		workouts:  deps.Workouts,
	}
	bootstrapSvc := &bootstrapService{
		users:      userSvc,
		bootstrap:  deps.Bootstrap,
		moderation: deps.Moderation,
		challenge:  deps.Challenge,
	}

	return &Service{
		Users:     userSvc,
		Circle:    circleSvc,
		Challenge: challengeSvc,
		Bootstrap: bootstrapSvc,
		rules:     rules,
	}
}

func normalizeRules(r Rules) Rules {
	if r.MinCircleDurationSeconds == 0 {
		r.MinCircleDurationSeconds = MinCircleDuration
	}
	if r.SessionGap == 0 {
		r.SessionGap = time.Duration(SessionGapMinutes) * time.Minute
	}
	return r
}

func (s *Service) RegisterUser(ctx context.Context, telegramID int64, username string) error {
	return s.Users.RegisterUser(ctx, telegramID, username)
}

func (s *Service) HandleCircle(ctx context.Context, userID int64, duration int, chatID int64, messageID int) error {
	return s.Circle.HandleCircle(ctx, userID, duration, chatID, messageID)
}

func (s *Service) CancelSession(ctx context.Context, chatID int64, messageID int) {
	s.Circle.CancelSession(ctx, chatID, messageID)
}

// StartChallenge deactivates current challenge and starts a new one.
func (s *Service) StartChallenge(ctx context.Context, chatID int64, daysPerWeek int, duration int) error {
	return s.Challenge.StartChallenge(ctx, chatID, daysPerWeek, duration)
}

// DeactivateChallengeForChat clears the active challenge for a chat.
func (s *Service) DeactivateChallengeForChat(ctx context.Context, chatID int64) error {
	return s.Challenge.DeactivateChallengeForChat(ctx, chatID)
}

// ActiveChallenges returns all active challenges.
func (s *Service) ActiveChallenges(ctx context.Context) ([]domain.Challenge, error) {
	return s.Challenge.ActiveChallenges(ctx)
}

func (s *Service) WeeklyCheck(ctx context.Context, challenge domain.Challenge) ([]UserInfo, error) {
	return s.Challenge.WeeklyCheck(ctx, challenge)
}

func (s *Service) InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
	return s.Bootstrap.InitChallengeBootstrap(ctx, chatID, welcomeMessageID, isBotAdmin, expectedReactions)
}

func (s *Service) UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error {
	return s.Bootstrap.UpsertChatMember(ctx, chatID, userID, isBot, isActive)
}

func (s *Service) SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error {
	return s.Bootstrap.SetBotAdminStatus(ctx, chatID, isBotAdmin)
}

func (s *Service) ProcessReactionUpdate(ctx context.Context, chatID int64, messageID int, userID int64, username string, emojis []string) (challengeStarted bool, workoutCancelled bool, err error) {
	return s.Bootstrap.ProcessReactionUpdate(ctx, chatID, messageID, userID, username, emojis)
}

func (s *Service) TryStartChallengeIfReady(ctx context.Context, chatID int64) (bool, error) {
	return s.Bootstrap.TryStartChallengeIfReady(ctx, chatID)
}

func (s *Service) UpdateChallengeConfig(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
	return s.Bootstrap.UpdateConfig(ctx, chatID, daysPerWeek, durationDays, price)
}

func (s *Service) GetChallengeConfig(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
	return s.Bootstrap.GetConfig(ctx, chatID)
}

func (s *Service) GetRules() Rules {
	return s.rules
}
