package domain

import (
	"time"
)

// User represents a bot participant.
type User struct {
	TelegramID  int64
	Username    string
	DaysTrained int
	IsActive    bool
	FailedAt    *time.Time
}

// Session tracks an in-progress training session.
type Session struct {
	ID          int
	UserID      int64
	ChatID      int64
	MessageID   int
	StartedAt   time.Time
	LastVideoAt time.Time
}

// Workout represents a completed workout.
type Workout struct {
	WorkoutDate time.Time
	ID          int
	UserID      int64
	ChatID      int64
	MessageID   int
	IsCancelled bool
	CancelledBy *int64
	CancelledAt *time.Time
}

// Vote represents a poll-based subtraction/addition process.
type Vote struct {
	ID           int
	ChatID       int64
	TargetUserID int64
	InitiatorID  int64
	PollID       string
	Amount       int
	Type         string // "subtract" or "add"
	CreatedAt    time.Time
	ExpiresAt    time.Time
	IsCompleted  bool
	IsSuccess    bool
}

// UserStats represents workout statistics for a user.
type UserStats struct {
	TelegramID      int64
	Username        string
	WeeklyCount     int
	TotalCount      int
	IsActive        bool
	HasWorkoutToday bool
}

// Challenge is the active challenge context.
type Challenge struct {
	ChallengeID       int
	IsActive          bool
	DaysPerWeek       int
	Duration          int
	Price             int
	StartedAt         time.Time
	ChatID            int64
	LastWeeklyCheckAt *time.Time
	LastDailyStatsAt  *time.Time
}

// ChallengeBootstrap stores the pre-start bootstrap state.
type ChallengeBootstrap struct {
	ChatID            int64
	WelcomeMessageID  int
	ExpectedReactions int
	RosterFrozenAt    time.Time
	IsStarted         bool
	StartedAt         *time.Time
	IsBotAdmin        bool
	DaysPerWeek       int
	DurationDays      int
	Price             int
}
