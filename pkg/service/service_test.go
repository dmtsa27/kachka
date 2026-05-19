package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

type fakeUsers struct {
	createUser       func(ctx context.Context, user domain.User) error
	readUser         func(ctx context.Context, telegramID int64) (*domain.User, error)
	getAllActive     func(ctx context.Context) ([]domain.User, error)
	batchDeactivate  func(ctx context.Context, userIDs []int64) error
}

func (f *fakeUsers) CreateUser(ctx context.Context, user domain.User) error {
	if f.createUser == nil {
		return fmt.Errorf("unexpected CreateUser")
	}
	return f.createUser(ctx, user)
}

func (f *fakeUsers) ReadUser(ctx context.Context, telegramID int64) (*domain.User, error) {
	if f.readUser == nil {
		return nil, fmt.Errorf("unexpected ReadUser")
	}
	return f.readUser(ctx, telegramID)
}

func (f *fakeUsers) GetAllActiveUsers(ctx context.Context) ([]domain.User, error) {
	if f.getAllActive == nil {
		return nil, fmt.Errorf("unexpected GetAllActiveUsers")
	}
	return f.getAllActive(ctx)
}

func (f *fakeUsers) BatchDeactivateUsers(ctx context.Context, userIDs []int64) error {
	if f.batchDeactivate == nil {
		return fmt.Errorf("unexpected BatchDeactivateUsers")
	}
	return f.batchDeactivate(ctx, userIDs)
}

type fakeSessions struct {
	hasTrainedToday  func(ctx context.Context, userID int64, chatID int64) (bool, error)
	startSession     func(ctx context.Context, userID int64, chatID int64, messageID int) error
	getSession       func(ctx context.Context, userID int64, chatID int64) (*domain.Session, error)
	addLatestSession func(ctx context.Context, userID int64, chatID int64) error
	deleteSession    func(ctx context.Context, chatID int64, messageID int) error
}

func (f *fakeSessions) HasTrainedToday(ctx context.Context, userID int64, chatID int64) (bool, error) {
	if f.hasTrainedToday == nil {
		return false, fmt.Errorf("unexpected HasTrainedToday")
	}
	return f.hasTrainedToday(ctx, userID, chatID)
}

func (f *fakeSessions) StartSession(ctx context.Context, userID int64, chatID int64, messageID int) error {
	if f.startSession == nil {
		return fmt.Errorf("unexpected StartSession")
	}
	return f.startSession(ctx, userID, chatID, messageID)
}

func (f *fakeSessions) GetSession(ctx context.Context, userID int64, chatID int64) (*domain.Session, error) {
	if f.getSession == nil {
		return nil, fmt.Errorf("unexpected GetSession")
	}
	return f.getSession(ctx, userID, chatID)
}

func (f *fakeSessions) AddLatestSession(ctx context.Context, userID int64, chatID int64) error {
	if f.addLatestSession == nil {
		return fmt.Errorf("unexpected AddLatestSession")
	}
	return f.addLatestSession(ctx, userID, chatID)
}

func (f *fakeSessions) DeleteSessionToday(ctx context.Context, chatID int64, messageID int) error {
	if f.deleteSession == nil {
		return fmt.Errorf("unexpected DeleteSessionToday")
	}
	return f.deleteSession(ctx, chatID, messageID)
}

type fakeWorkouts struct {
	hasWorkoutToday func(ctx context.Context, userID int64, chatID int64) (bool, error)
	createWorkout   func(ctx context.Context, workout domain.Workout) error
	weeklyWorkouts  func(ctx context.Context, userID int64, weekStart time.Time) (int, error)
	getWorkoutCounts func(ctx context.Context, weekStart time.Time) ([]UserWorkouts, error)
}

func (f *fakeWorkouts) HasWorkoutToday(ctx context.Context, userID int64, chatID int64) (bool, error) {
	if f.hasWorkoutToday == nil {
		return false, fmt.Errorf("unexpected HasWorkoutToday")
	}
	return f.hasWorkoutToday(ctx, userID, chatID)
}

func (f *fakeWorkouts) CreateWorkout(ctx context.Context, workout domain.Workout) error {
	if f.createWorkout == nil {
		return fmt.Errorf("unexpected CreateWorkout")
	}
	return f.createWorkout(ctx, workout)
}

func (f *fakeWorkouts) WeeklyWorkouts(ctx context.Context, userID int64, weekStart time.Time) (int, error) {
	if f.weeklyWorkouts == nil {
		return 0, fmt.Errorf("unexpected WeeklyWorkouts")
	}
	return f.weeklyWorkouts(ctx, userID, weekStart)
}

func (f *fakeWorkouts) GetWorkoutCounts(ctx context.Context, weekStart time.Time) ([]UserWorkouts, error) {
	if f.getWorkoutCounts == nil {
		return nil, fmt.Errorf("unexpected GetWorkoutCounts")
	}
	return f.getWorkoutCounts(ctx, weekStart)
}

type fakeChallenges struct {
	getActiveChallenge     func(ctx context.Context) (*domain.Challenge, error)
	getAllActiveChallenges func(ctx context.Context) ([]domain.Challenge, error)
	hasActiveChallengeChat func(ctx context.Context, chatID int64) (bool, error)
	createChallenge        func(ctx context.Context, challenge domain.Challenge) error
	deactivateForChat      func(ctx context.Context, chatID int64) error
}

func (f *fakeChallenges) GetActiveChallenge(ctx context.Context) (*domain.Challenge, error) {
	if f.getActiveChallenge == nil {
		return nil, fmt.Errorf("unexpected GetActiveChallenge")
	}
	return f.getActiveChallenge(ctx)
}

func (f *fakeChallenges) GetAllActiveChallenges(ctx context.Context) ([]domain.Challenge, error) {
	if f.getAllActiveChallenges == nil {
		return nil, fmt.Errorf("unexpected GetAllActiveChallenges")
	}
	return f.getAllActiveChallenges(ctx)
}

func (f *fakeChallenges) HasActiveChallengeInChat(ctx context.Context, chatID int64) (bool, error) {
	if f.hasActiveChallengeChat == nil {
		return false, fmt.Errorf("unexpected HasActiveChallengeInChat")
	}
	return f.hasActiveChallengeChat(ctx, chatID)
}

func (f *fakeChallenges) CreateChallenge(ctx context.Context, challenge domain.Challenge) error {
	if f.createChallenge == nil {
		return fmt.Errorf("unexpected CreateChallenge")
	}
	return f.createChallenge(ctx, challenge)
}

func (f *fakeChallenges) DeactivateChallengeForChat(ctx context.Context, chatID int64) error {
	if f.deactivateForChat == nil {
		return fmt.Errorf("unexpected DeactivateChallengeForChat")
	}
	return f.deactivateForChat(ctx, chatID)
}

type fakeBootstrap struct {
	initBootstrap       func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error
	getBootstrap        func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error)
	setBotAdmin         func(ctx context.Context, chatID int64, isBotAdmin bool) error
	upsertChatMember    func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error
	setReactions        func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error
	countHeartReactions func(ctx context.Context, chatID int64) (int, error)
	markStarted         func(ctx context.Context, chatID int64) (bool, error)
	updateConfig        func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error
}

func (f *fakeBootstrap) InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
	if f.initBootstrap == nil {
		return fmt.Errorf("unexpected InitChallengeBootstrap")
	}
	return f.initBootstrap(ctx, chatID, welcomeMessageID, isBotAdmin, expectedReactions)
}

func (f *fakeBootstrap) GetChallengeBootstrap(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
	if f.getBootstrap == nil {
		return nil, fmt.Errorf("unexpected GetChallengeBootstrap")
	}
	return f.getBootstrap(ctx, chatID)
}

func (f *fakeBootstrap) SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error {
	if f.setBotAdmin == nil {
		return fmt.Errorf("unexpected SetBotAdminStatus")
	}
	return f.setBotAdmin(ctx, chatID, isBotAdmin)
}

func (f *fakeBootstrap) UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error {
	if f.upsertChatMember == nil {
		return fmt.Errorf("unexpected UpsertChatMember")
	}
	return f.upsertChatMember(ctx, chatID, userID, isBot, isActive)
}

func (f *fakeBootstrap) SetUserMessageReactions(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
	if f.setReactions == nil {
		return fmt.Errorf("unexpected SetUserMessageReactions")
	}
	return f.setReactions(ctx, chatID, messageID, userID, emojis)
}

func (f *fakeBootstrap) CountWelcomeHeartReactions(ctx context.Context, chatID int64) (int, error) {
	if f.countHeartReactions == nil {
		return 0, fmt.Errorf("unexpected CountWelcomeHeartReactions")
	}
	return f.countHeartReactions(ctx, chatID)
}

func (f *fakeBootstrap) MarkChallengeStarted(ctx context.Context, chatID int64) (bool, error) {
	if f.markStarted == nil {
		return false, fmt.Errorf("unexpected MarkChallengeStarted")
	}
	return f.markStarted(ctx, chatID)
}

func (f *fakeBootstrap) UpdateChallengeBootstrapConfig(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
	if f.updateConfig == nil {
		return fmt.Errorf("unexpected UpdateChallengeBootstrapConfig")
	}
	return f.updateConfig(ctx, chatID, daysPerWeek, durationDays, price)
}

type fakeModeration struct {
	cancelCounted func(ctx context.Context, chatID int64, messageID int) (bool, error)
}

func (f *fakeModeration) CancelCountedByMessage(ctx context.Context, chatID int64, messageID int) (bool, error) {
	if f.cancelCounted == nil {
		return false, fmt.Errorf("unexpected CancelCountedByMessage")
	}
	return f.cancelCounted(ctx, chatID, messageID)
}

type fakeNotifier struct {
	messages []string
}

func (f *fakeNotifier) SendMessage(ctx context.Context, chatID int64, text string) error {
	f.messages = append(f.messages, text)
	return nil
}

func TestHandleCircle_StartsSession(t *testing.T) {
	ctx := context.Background()
	notifier := &fakeNotifier{}

	users := &fakeUsers{
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
			return &domain.User{TelegramID: telegramID, IsActive: true}, nil
		},
		createUser: func(ctx context.Context, user domain.User) error {
			return nil
		},
		getAllActive: func(ctx context.Context) ([]domain.User, error) {
			return nil, nil
		},
		batchDeactivate: func(ctx context.Context, userIDs []int64) error {
			return nil
		},
	}

	var started bool
	var latestAdded bool
	var createdWorkout bool

	sessions := &fakeSessions{
		hasTrainedToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) {
			return false, nil
		},
		startSession: func(ctx context.Context, userID int64, chatID int64, messageID int) error {
			started = true
			return nil
		},
		addLatestSession: func(ctx context.Context, userID int64, chatID int64) error {
			latestAdded = true
			return nil
		},
		getSession: func(ctx context.Context, userID int64, chatID int64) (*domain.Session, error) {
			return &domain.Session{}, nil
		},
		deleteSession: func(ctx context.Context, chatID int64, messageID int) error {
			return nil
		},
	}

	workouts := &fakeWorkouts{
		hasWorkoutToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) {
			return false, nil
		},
		createWorkout: func(ctx context.Context, workout domain.Workout) error {
			createdWorkout = true
			return nil
		},
		weeklyWorkouts: func(ctx context.Context, userID int64, weekStart time.Time) (int, error) {
			return 0, nil
		},
	}

	challenges := &fakeChallenges{
		hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) {
			return true, nil
		},
		getActiveChallenge: func(ctx context.Context) (*domain.Challenge, error) {
			return nil, errors.New("unexpected")
		},
		createChallenge: func(ctx context.Context, challenge domain.Challenge) error {
			return nil
		},
		deactivateForChat: func(ctx context.Context, chatID int64) error {
			return nil
		},
	}

	bootstrap := &fakeBootstrap{
		initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
			return nil
		},
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
			return nil, sql.ErrNoRows
		},
		setBotAdmin: func(ctx context.Context, chatID int64, isBotAdmin bool) error {
			return nil
		},
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error {
			return nil
		},
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
			return nil
		},
		countHeartReactions: func(ctx context.Context, chatID int64) (int, error) {
			return 0, nil
		},
		markStarted: func(ctx context.Context, chatID int64) (bool, error) {
			return false, nil
		},
		updateConfig: func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
			return nil
		},
	}

	moderation := &fakeModeration{
		cancelCounted: func(ctx context.Context, chatID int64, messageID int) (bool, error) {
			return false, nil
		},
	}

	svc := New(Deps{
		Users:      users,
		Sessions:   sessions,
		Workouts:   workouts,
		Challenge:  challenges,
		Bootstrap:  bootstrap,
		Moderation: moderation,
		Notifier:   notifier,
	})

	if err := svc.HandleCircle(ctx, 10, MinCircleDuration, 20, 30); err != nil {
		t.Fatalf("HandleCircle error: %v", err)
	}

	if !started {
		t.Fatalf("expected StartSession")
	}
	if latestAdded {
		t.Fatalf("did not expect AddLatestSession")
	}
	if createdWorkout {
		t.Fatalf("did not expect CreateWorkout")
	}
	if len(notifier.messages) != 1 || notifier.messages[0] != "🚀" {
		t.Fatalf("expected rocket notification")
	}
}

func TestHandleCircle_CompletesWorkout(t *testing.T) {
	ctx := context.Background()
	notifier := &fakeNotifier{}

	users := &fakeUsers{
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
			return &domain.User{TelegramID: telegramID, IsActive: true}, nil
		},
		createUser: func(ctx context.Context, user domain.User) error { return nil },
		getAllActive: func(ctx context.Context) ([]domain.User, error) { return nil, nil },
		batchDeactivate: func(ctx context.Context, userIDs []int64) error { return nil },
	}

	var createdWorkout bool
	var latestAdded bool

	sessions := &fakeSessions{
		hasTrainedToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) {
			return true, nil
		},
		startSession: func(ctx context.Context, userID int64, chatID int64, messageID int) error {
			return nil
		},
		addLatestSession: func(ctx context.Context, userID int64, chatID int64) error {
			latestAdded = true
			return nil
		},
		getSession: func(ctx context.Context, userID int64, chatID int64) (*domain.Session, error) {
			startedAt := time.Now().Add(-30 * time.Minute)
			return &domain.Session{StartedAt: startedAt, LastVideoAt: time.Now()}, nil
		},
		deleteSession: func(ctx context.Context, chatID int64, messageID int) error { return nil },
	}

	workouts := &fakeWorkouts{
		hasWorkoutToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) { return false, nil },
		createWorkout: func(ctx context.Context, workout domain.Workout) error {
			createdWorkout = true
			return nil
		},
		weeklyWorkouts: func(ctx context.Context, userID int64, weekStart time.Time) (int, error) { return 0, nil },
	}

	challenges := &fakeChallenges{
		hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) { return true, nil },
		getActiveChallenge: func(ctx context.Context) (*domain.Challenge, error) { return nil, errors.New("unexpected") },
		createChallenge: func(ctx context.Context, challenge domain.Challenge) error { return nil },
		deactivateForChat: func(ctx context.Context, chatID int64) error { return nil },
	}

	bootstrap := &fakeBootstrap{
		initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error { return nil },
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) { return nil, sql.ErrNoRows },
		setBotAdmin: func(ctx context.Context, chatID int64, isBotAdmin bool) error { return nil },
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error { return nil },
		countHeartReactions: func(ctx context.Context, chatID int64) (int, error) { return 0, nil },
		markStarted: func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
		updateConfig: func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error { return nil },
	}

	moderation := &fakeModeration{cancelCounted: func(ctx context.Context, chatID int64, messageID int) (bool, error) { return false, nil }}

	svc := New(Deps{
		Users:      users,
		Sessions:   sessions,
		Workouts:   workouts,
		Challenge:  challenges,
		Bootstrap:  bootstrap,
		Moderation: moderation,
		Notifier:   notifier,
	})

	if err := svc.HandleCircle(ctx, 10, MinCircleDuration, 20, 30); err != nil {
		t.Fatalf("HandleCircle error: %v", err)
	}

	if !latestAdded {
		t.Fatalf("expected AddLatestSession")
	}
	if !createdWorkout {
		t.Fatalf("expected CreateWorkout")
	}
	if len(notifier.messages) != 1 || notifier.messages[0] != "🎖" {
		t.Fatalf("expected medal notification")
	}
}

func TestProcessReactionUpdate_CancelsWorkout(t *testing.T) {
	ctx := context.Background()

	users := &fakeUsers{
		createUser: func(ctx context.Context, user domain.User) error { return nil },
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) { return &domain.User{}, nil },
		getAllActive: func(ctx context.Context) ([]domain.User, error) { return nil, nil },
		batchDeactivate: func(ctx context.Context, userIDs []int64) error { return nil },
	}

	bootstrap := &fakeBootstrap{
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error { return nil },
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) { return nil, sql.ErrNoRows },
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
		initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error { return nil },
		setBotAdmin: func(ctx context.Context, chatID int64, isBotAdmin bool) error { return nil },
		countHeartReactions: func(ctx context.Context, chatID int64) (int, error) { return 0, nil },
		markStarted: func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
		updateConfig: func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error { return nil },
	}

	moderation := &fakeModeration{
		cancelCounted: func(ctx context.Context, chatID int64, messageID int) (bool, error) { return true, nil },
	}

	svc := New(Deps{
		Users:      users,
		Sessions:   &fakeSessions{},
		Workouts:   &fakeWorkouts{},
		Challenge:  &fakeChallenges{hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) { return false, nil }, createChallenge: func(ctx context.Context, challenge domain.Challenge) error { return nil }},
		Bootstrap:  bootstrap,
		Moderation: moderation,
	})

	started, cancelled, err := svc.ProcessReactionUpdate(ctx, 1, 2, 3, "u", []string{"👎"})
	if err != nil {
		t.Fatalf("ProcessReactionUpdate error: %v", err)
	}
	if started {
		t.Fatalf("did not expect challenge to start")
	}
	if !cancelled {
		t.Fatalf("expected workout to be cancelled")
	}
}

func TestTryStartChallengeIfReady_Starts(t *testing.T) {
	ctx := context.Background()

	bootstrap := &fakeBootstrap{
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
			return &domain.ChallengeBootstrap{ChatID: chatID, ExpectedReactions: 2, IsBotAdmin: true, DaysPerWeek: 3, DurationDays: 180}, nil
		},
		countHeartReactions: func(ctx context.Context, chatID int64) (int, error) { return 2, nil },
		markStarted: func(ctx context.Context, chatID int64) (bool, error) { return true, nil },
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error { return nil },
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
		initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error { return nil },
		setBotAdmin: func(ctx context.Context, chatID int64, isBotAdmin bool) error { return nil },
	}

	created := false
	challenges := &fakeChallenges{
		createChallenge: func(ctx context.Context, challenge domain.Challenge) error {
			created = true
			return nil
		},
		getActiveChallenge: func(ctx context.Context) (*domain.Challenge, error) { return nil, errors.New("unexpected") },
		deactivateForChat: func(ctx context.Context, chatID int64) error { return nil },
		hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
	}

	svc := New(Deps{
		Users:      &fakeUsers{createUser: func(ctx context.Context, user domain.User) error { return nil }, readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) { return &domain.User{}, nil }, getAllActive: func(ctx context.Context) ([]domain.User, error) { return nil, nil }, batchDeactivate: func(ctx context.Context, userIDs []int64) error { return nil }},
		Sessions:   &fakeSessions{},
		Workouts:   &fakeWorkouts{},
		Challenge:  challenges,
		Bootstrap:  bootstrap,
		Moderation: &fakeModeration{cancelCounted: func(ctx context.Context, chatID int64, messageID int) (bool, error) { return false, nil }},
	})

	started, err := svc.TryStartChallengeIfReady(ctx, 100)
	if err != nil {
		t.Fatalf("TryStartChallengeIfReady error: %v", err)
	}
	if !started {
		t.Fatalf("expected challenge to start")
	}
	if !created {
		t.Fatalf("expected CreateChallenge to be called")
	}
}

func TestWeeklyCheck_DeactivatesFailed(t *testing.T) {
	ctx := context.Background()

	startedAt := time.Now().Add(-8 * 24 * time.Hour)
	challenge := &domain.Challenge{ChatID: 1, DaysPerWeek: 3, StartedAt: startedAt}

	challenges := &fakeChallenges{
		getActiveChallenge: func(ctx context.Context) (*domain.Challenge, error) { return challenge, nil },
		createChallenge: func(ctx context.Context, challenge domain.Challenge) error { return nil },
		deactivateForChat: func(ctx context.Context, chatID int64) error { return nil },
		hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
	}

	users := &fakeUsers{
		createUser: func(ctx context.Context, user domain.User) error { return nil },
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) { return &domain.User{}, nil },
		getAllActive: func(ctx context.Context) ([]domain.User, error) {
			return []domain.User{{TelegramID: 10, Username: "a"}, {TelegramID: 11, Username: "b"}}, nil
		},
		batchDeactivate: func(ctx context.Context, userIDs []int64) error {
			if len(userIDs) != 1 || userIDs[0] != 10 {
				return fmt.Errorf("unexpected deactivation list")
			}
			return nil
		},
	}

	workouts := &fakeWorkouts{
		hasWorkoutToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) { return false, nil },
		createWorkout: func(ctx context.Context, workout domain.Workout) error { return nil },
		getWorkoutCounts: func(ctx context.Context, weekStart time.Time) ([]UserWorkouts, error) {
			if weekStart.Before(startedAt.Add(7 * 24 * time.Hour)) {
				return nil, fmt.Errorf("unexpected week start")
			}
			return []UserWorkouts{
				{UserInfo: UserInfo{TelegramID: 10, Username: "a"}, Count: 2},
				{UserInfo: UserInfo{TelegramID: 11, Username: "b"}, Count: 3},
			}, nil
		},
	}

	svc := New(Deps{
		Users:      users,
		Sessions:   &fakeSessions{},
		Workouts:   workouts,
		Challenge:  challenges,
		Bootstrap:  &fakeBootstrap{
			initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error { return nil },
			getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) { return nil, sql.ErrNoRows },
			setBotAdmin: func(ctx context.Context, chatID int64, isBotAdmin bool) error { return nil },
			upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
			setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error { return nil },
			countHeartReactions: func(ctx context.Context, chatID int64) (int, error) { return 0, nil },
			markStarted: func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
			updateConfig: func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error { return nil },
		},
		Moderation: &fakeModeration{cancelCounted: func(ctx context.Context, chatID int64, messageID int) (bool, error) { return false, nil }},
	})

	failed, err := svc.WeeklyCheck(ctx, *challenge)
	if err != nil {
		t.Fatalf("WeeklyCheck error: %v", err)
	}
	if len(failed) != 1 || failed[0].TelegramID != 10 {
		t.Fatalf("unexpected failed list")
	}
}
