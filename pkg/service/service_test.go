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
	createUser          func(ctx context.Context, user domain.User) error
	readUser            func(ctx context.Context, telegramID int64) (*domain.User, error)
	getAllActive        func(ctx context.Context) ([]domain.User, error)
	batchDeactivate     func(ctx context.Context, userIDs []int64) error
	getUserIDByUsername func(ctx context.Context, username string) (int64, error)
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

func (f *fakeUsers) GetUserIDByUsername(ctx context.Context, username string) (int64, error) {
	if f.getUserIDByUsername == nil {
		return 0, fmt.Errorf("unexpected GetUserIDByUsername")
	}
	return f.getUserIDByUsername(ctx, username)
}

type fakeSessions struct {
	hasTrainedToday     func(ctx context.Context, userID int64, chatID int64) (bool, error)
	startSession        func(ctx context.Context, userID int64, chatID int64, messageID int) error
	getSession          func(ctx context.Context, userID int64, chatID int64) (*domain.Session, error)
	getSessionByMessage func(ctx context.Context, chatID int64, messageID int) (*domain.Session, error)
	addLatestSession    func(ctx context.Context, userID int64, chatID int64) error
	deleteSession       func(ctx context.Context, chatID int64, messageID int) error
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

func (f *fakeSessions) GetSessionByMessage(ctx context.Context, chatID int64, messageID int) (*domain.Session, error) {
	if f.getSessionByMessage == nil {
		return nil, fmt.Errorf("unexpected GetSessionByMessage")
	}
	return f.getSessionByMessage(ctx, chatID, messageID)
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
	hasWorkoutToday  func(ctx context.Context, userID int64, chatID int64) (bool, error)
	createWorkout    func(ctx context.Context, workout domain.Workout) error
	weeklyWorkouts   func(ctx context.Context, userID int64, weekStart time.Time) (int, error)
	getWorkoutCounts func(ctx context.Context, weekStart time.Time) ([]UserWorkouts, error)
	cancelWorkout    func(ctx context.Context, chatID int64, messageID int, cancelledBy int64) (int64, error)
	reinstateWorkout func(ctx context.Context, chatID int64, messageID int, reinstatedBy int64) error
	getWorkoutByMsg  func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error)
	subtractWorkouts func(ctx context.Context, userID int64, chatID int64, amount int) (int, error)
	addWorkouts      func(ctx context.Context, userID int64, chatID int64, amount int) (int, error)
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

func (f *fakeWorkouts) CancelWorkout(ctx context.Context, chatID int64, messageID int, cancelledBy int64) (int64, error) {
	if f.cancelWorkout == nil {
		return 0, fmt.Errorf("unexpected CancelWorkout")
	}
	return f.cancelWorkout(ctx, chatID, messageID, cancelledBy)
}

func (f *fakeWorkouts) ReinstateWorkout(ctx context.Context, chatID int64, messageID int, reinstatedBy int64) error {
	if f.reinstateWorkout == nil {
		return fmt.Errorf("unexpected ReinstateWorkout")
	}
	return f.reinstateWorkout(ctx, chatID, messageID, reinstatedBy)
}

func (f *fakeWorkouts) GetWorkoutByMessage(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
	if f.getWorkoutByMsg == nil {
		return nil, fmt.Errorf("unexpected GetWorkoutByMessage")
	}
	return f.getWorkoutByMsg(ctx, chatID, messageID)
}

func (f *fakeWorkouts) SubtractWorkouts(ctx context.Context, userID int64, chatID int64, amount int) (int, error) {
	if f.subtractWorkouts == nil {
		return 0, fmt.Errorf("unexpected SubtractWorkouts")
	}
	return f.subtractWorkouts(ctx, userID, chatID, amount)
}

func (f *fakeWorkouts) AddWorkouts(ctx context.Context, userID int64, chatID int64, amount int) (int, error) {
	if f.addWorkouts == nil {
		return 0, fmt.Errorf("unexpected AddWorkouts")
	}
	return f.addWorkouts(ctx, userID, chatID, amount)
}

type fakeChallenges struct {
	getActiveChallenge     func(ctx context.Context) (*domain.Challenge, error)
	getActiveByChat        func(ctx context.Context, chatID int64) (*domain.Challenge, error)
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

func (f *fakeChallenges) GetActiveChallengeByChat(ctx context.Context, chatID int64) (*domain.Challenge, error) {
	if f.getActiveByChat == nil {
		return nil, fmt.Errorf("unexpected GetActiveChallengeByChat")
	}
	return f.getActiveByChat(ctx, chatID)
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
	cancelCounted    func(ctx context.Context, chatID int64, messageID int) (bool, error)
	disputeWorkout   func(ctx context.Context, chatID int64, messageID int, disputerID int64) (int64, bool, error)
	reinstateWorkout func(ctx context.Context, chatID int64, messageID int, reinstaterID int64) error
	initiateSubtract func(ctx context.Context, chatID int64, initiatorID int64, targetUsername string, amount int, pollID string) error
	initiateAdd      func(ctx context.Context, chatID int64, initiatorID int64, targetUsername string, amount int, pollID string) error
	handlePollUpdate func(ctx context.Context, pollID string, success bool) (int, error)
	getWorkoutByMsg  func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error)
}

func (f *fakeModeration) CancelCountedByMessage(ctx context.Context, chatID int64, messageID int) (bool, error) {
	if f.cancelCounted == nil {
		return false, fmt.Errorf("unexpected CancelCountedByMessage")
	}
	return f.cancelCounted(ctx, chatID, messageID)
}

func (f *fakeModeration) DisputeWorkout(ctx context.Context, chatID int64, messageID int, disputerID int64) (int64, bool, error) {
	if f.disputeWorkout == nil {
		return 0, false, fmt.Errorf("unexpected DisputeWorkout")
	}
	return f.disputeWorkout(ctx, chatID, messageID, disputerID)
}

func (f *fakeModeration) ReinstateWorkout(ctx context.Context, chatID int64, messageID int, reinstaterID int64) error {
	if f.reinstateWorkout == nil {
		return fmt.Errorf("unexpected ReinstateWorkout")
	}
	return f.reinstateWorkout(ctx, chatID, messageID, reinstaterID)
}

func (f *fakeModeration) InitiateSubtract(ctx context.Context, chatID int64, initiatorID int64, targetUsername string, amount int, pollID string) error {
	if f.initiateSubtract == nil {
		return fmt.Errorf("unexpected InitiateSubtract")
	}
	return f.initiateSubtract(ctx, chatID, initiatorID, targetUsername, amount, pollID)
}

func (f *fakeModeration) InitiateAdd(ctx context.Context, chatID int64, initiatorID int64, targetUsername string, amount int, pollID string) error {
	if f.initiateAdd == nil {
		return fmt.Errorf("unexpected InitiateAdd")
	}
	return f.initiateAdd(ctx, chatID, initiatorID, targetUsername, amount, pollID)
}

func (f *fakeModeration) HandlePollUpdate(ctx context.Context, pollID string, success bool) (int, error) {
	if f.handlePollUpdate == nil {
		return 0, fmt.Errorf("unexpected HandlePollUpdate")
	}
	return f.handlePollUpdate(ctx, pollID, success)
}

func (f *fakeModeration) GetWorkoutByMessage(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
	if f.getWorkoutByMsg == nil {
		return nil, fmt.Errorf("unexpected GetWorkoutByMessage")
	}
	return f.getWorkoutByMsg(ctx, chatID, messageID)
}

type fakeVotes struct {
	createVote      func(ctx context.Context, vote domain.Vote) error
	getVoteByPollID func(ctx context.Context, pollID string) (*domain.Vote, error)
	completeVote    func(ctx context.Context, pollID string, success bool) error
}

func (f *fakeVotes) CreateVote(ctx context.Context, vote domain.Vote) error {
	if f.createVote == nil {
		return fmt.Errorf("unexpected CreateVote")
	}
	return f.createVote(ctx, vote)
}

func (f *fakeVotes) GetVoteByPollID(ctx context.Context, pollID string) (*domain.Vote, error) {
	if f.getVoteByPollID == nil {
		return nil, fmt.Errorf("unexpected GetVoteByPollID")
	}
	return f.getVoteByPollID(ctx, pollID)
}

func (f *fakeVotes) CompleteVote(ctx context.Context, pollID string, success bool) error {
	if f.completeVote == nil {
		return fmt.Errorf("unexpected CompleteVote")
	}
	return f.completeVote(ctx, pollID, success)
}

type fakeNotifier struct {
	messages []string
}

func (f *fakeNotifier) SendMessage(ctx context.Context, chatID int64, text string) (int, error) {
	f.messages = append(f.messages, text)
	return 123, nil // Return a fake message ID
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

func TestBootstrapService_Config(t *testing.T) {
	ctx := context.Background()
	bootstrap := &fakeBootstrap{
		initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
			return nil
		},
		updateConfig: func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
			return nil
		},
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
			return &domain.ChallengeBootstrap{DaysPerWeek: 5}, nil
		},
		setBotAdmin: func(ctx context.Context, chatID int64, isBotAdmin bool) error {
			return nil
		},
	}

	svc := New(Deps{
		Bootstrap: bootstrap,
	})

	if err := svc.InitChallengeBootstrap(ctx, 1, 2, true, 3); err != nil {
		t.Errorf("InitChallengeBootstrap error: %v", err)
	}

	if err := svc.SetBotAdminStatus(ctx, 1, true); err != nil {
		t.Errorf("SetBotAdminStatus error: %v", err)
	}

	if err := svc.UpdateChallengeConfig(ctx, 1, 5, 30, 500); err != nil {
		t.Errorf("UpdateChallengeConfig error: %v", err)
	}

	config, err := svc.GetChallengeConfig(ctx, 1)
	if err != nil {
		t.Errorf("GetChallengeConfig error: %v", err)
	}
	if config.DaysPerWeek != 5 {
		t.Errorf("expected DaysPerWeek 5, got %d", config.DaysPerWeek)
	}
}

func TestChallengeService_Management(t *testing.T) {
	ctx := context.Background()
	challenges := &fakeChallenges{
		createChallenge: func(ctx context.Context, challenge domain.Challenge) error {
			return nil
		},
		deactivateForChat: func(ctx context.Context, chatID int64) error {
			return nil
		},
	}

	svc := New(Deps{
		Challenge: challenges,
	})

	if err := svc.StartChallenge(ctx, 1, 3, 30); err != nil {
		t.Errorf("StartChallenge error: %v", err)
	}

	if err := svc.DeactivateChallengeForChat(ctx, 1); err != nil {
		t.Errorf("DeactivateChallengeForChat error: %v", err)
	}
}

func TestCircleService_Cancel(t *testing.T) {
	ctx := context.Background()
	sessions := &fakeSessions{
		deleteSession: func(ctx context.Context, chatID int64, messageID int) error {
			return nil
		},
	}

	svc := New(Deps{
		Sessions: sessions,
	})

	svc.CancelSession(ctx, 1, 2)
}

func TestUserService_IsActiveUser(t *testing.T) {
	ctx := context.Background()
	users := &fakeUsers{
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
			if telegramID == 1 {
				return &domain.User{IsActive: true}, nil
			}
			return nil, sql.ErrNoRows
		},
	}

	svc := New(Deps{
		Users: users,
	})

	active, err := svc.Users.IsActiveUser(ctx, 1)
	if err != nil || !active {
		t.Errorf("expected active user 1")
	}

	active, err = svc.Users.IsActiveUser(ctx, 2)
	if err != nil || active {
		t.Errorf("expected inactive user 2")
	}
}

func TestProcessReactionUpdate_EdgeCases(t *testing.T) {
	ctx := context.Background()

	bootstrap := &fakeBootstrap{
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
			if chatID == 999 {
				return nil, sql.ErrNoRows
			}
			return &domain.ChallengeBootstrap{WelcomeMessageID: 100}, nil
		},
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
			return nil
		},
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error {
			return nil
		},
	}

	svc := New(Deps{
		Users: &fakeUsers{createUser: func(ctx context.Context, user domain.User) error { return nil }},
		Bootstrap: bootstrap,
		Moderation: &fakeModeration{cancelCounted: func(ctx context.Context, chatID int64, messageID int) (bool, error) { return false, nil }},
	})

	// Case 1: Wrong message ID
	started, cancelled, err := svc.ProcessReactionUpdate(ctx, 1, 200, 3, "u", []string{"❤"})
	if err != nil || started || cancelled {
		t.Errorf("expected no action for wrong message ID")
	}

	// Case 2: Bootstrap not found
	started, cancelled, err = svc.ProcessReactionUpdate(ctx, 999, 100, 3, "u", []string{"❤"})
	if err != nil || started || cancelled {
		t.Errorf("expected no action for missing bootstrap")
	}
}

func TestTryStartChallengeIfReady_EdgeCases(t *testing.T) {
	ctx := context.Background()

	bootstrap := &fakeBootstrap{
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
			if chatID == 1 { // Not admin
				return &domain.ChallengeBootstrap{IsBotAdmin: false}, nil
			}
			if chatID == 2 { // Already started
				return &domain.ChallengeBootstrap{IsBotAdmin: true, IsStarted: true}, nil
			}
			if chatID == 3 { // Expected reactions <= 0
				return &domain.ChallengeBootstrap{IsBotAdmin: true, ExpectedReactions: 0}, nil
			}
			if chatID == 4 { // Not enough hearts
				return &domain.ChallengeBootstrap{IsBotAdmin: true, ExpectedReactions: 5}, nil
			}
			return nil, sql.ErrNoRows
		},
		countHeartReactions: func(ctx context.Context, chatID int64) (int, error) {
			return 2, nil
		},
	}

	svc := New(Deps{
		Bootstrap: bootstrap,
	})

	for i := 1; i <= 4; i++ {
		started, err := svc.TryStartChallengeIfReady(ctx, int64(i))
		if err != nil || started {
			t.Errorf("expected challenge NOT to start for case %d", i)
		}
	}

	started, err := svc.TryStartChallengeIfReady(ctx, 999)
	if err != nil || started {
		t.Errorf("expected challenge NOT to start for missing bootstrap")
	}
}

func TestHandleCircle_EdgeCases(t *testing.T) {
	ctx := context.Background()

	svc := New(Deps{
		Challenge: &fakeChallenges{
			hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) {
				return chatID == 1, nil
			},
		},
		Users: &fakeUsers{
			readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
				return &domain.User{IsActive: telegramID == 10}, nil
			},
		},
		Workouts: &fakeWorkouts{
			hasWorkoutToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) {
				return userID == 11, nil
			},
		},
		Rules: Rules{MinCircleDurationSeconds: 10},
	})

	// Case 1: Duration too short
	if err := svc.HandleCircle(ctx, 10, 5, 1, 1); err != nil {
		t.Errorf("HandleCircle error: %v", err)
	}

	// Case 2: Inactive chat
	if err := svc.HandleCircle(ctx, 10, 15, 2, 1); err != nil {
		t.Errorf("HandleCircle error: %v", err)
	}

	// Case 3: Inactive user
	if err := svc.HandleCircle(ctx, 99, 15, 1, 1); err != nil {
		t.Errorf("HandleCircle error: %v", err)
	}

	// Case 4: Already has workout
	if err := svc.HandleCircle(ctx, 11, 15, 1, 1); err != nil {
		t.Errorf("HandleCircle error: %v", err)
	}
}

func TestService_DefaultRules(t *testing.T) {
	rules := DefaultRules()
	if rules.MinCircleDurationSeconds != MinCircleDuration {
		t.Errorf("expected default min duration %d, got %d", MinCircleDuration, rules.MinCircleDurationSeconds)
	}
}

func TestService_FacadeDelegation(t *testing.T) {
	ctx := context.Background()
	
	// Testing delegation of methods that were at 0% in service.go
	svc := New(Deps{
		Bootstrap: &fakeBootstrap{
			upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
		},
		Challenge: &fakeChallenges{
			getAllActiveChallenges: func(ctx context.Context) ([]domain.Challenge, error) { return nil, nil },
		},
	})
	
	_ = svc.UpsertChatMember(ctx, 1, 2, false, true)
	_, _ = svc.ActiveChallenges(ctx)
}

func TestBootstrapService_UpdateConfig_Started(t *testing.T) {
	ctx := context.Background()
	bootstrap := &fakeBootstrap{
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
			return &domain.ChallengeBootstrap{IsStarted: true}, nil
		},
	}

	svc := New(Deps{
		Bootstrap: bootstrap,
	})

	err := svc.UpdateChallengeConfig(ctx, 1, 5, 30, 500)
	if err == nil || err.Error() != "cannot change configuration after challenge has started or rules confirmed" {
		t.Errorf("expected error 'cannot change configuration after challenge has started or rules confirmed', got %v", err)
	}
}
