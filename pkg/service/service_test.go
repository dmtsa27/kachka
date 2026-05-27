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
	updateUser          func(ctx context.Context, user domain.User) error
	deleteUser          func(ctx context.Context, telegramID int64) error
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

func (f *fakeUsers) UpdateUser(ctx context.Context, user domain.User) error {
	if f.updateUser == nil {
		return fmt.Errorf("unexpected UpdateUser")
	}
	return f.updateUser(ctx, user)
}

func (f *fakeUsers) DeleteUser(ctx context.Context, telegramID int64) error {
	if f.deleteUser == nil {
		return fmt.Errorf("unexpected DeleteUser")
	}
	return f.deleteUser(ctx, telegramID)
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
	weeklyWorkouts   func(ctx context.Context, userID int64, chatID int64, weekStart time.Time) (int, error)
	getWorkoutCounts func(ctx context.Context, chatID int64, weekStart time.Time) ([]UserWorkouts, error)
	cancelWorkout    func(ctx context.Context, chatID int64, messageID int, cancelledBy int64) (int64, error)
	reinstateWorkout func(ctx context.Context, chatID int64, messageID int, reinstatedBy int64) error
	getWorkoutByMsg  func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error)
	subtractWorkouts func(ctx context.Context, userID int64, chatID int64, amount int) (int, error)
	addWorkouts      func(ctx context.Context, userID int64, chatID int64, amount int) (int, error)
	getChatStats     func(ctx context.Context, chatID int64, weekStart time.Time) ([]domain.UserStats, error)
	getVotersCount   func(ctx context.Context, chatID int64) (int, error)
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

func (f *fakeWorkouts) WeeklyWorkouts(ctx context.Context, userID int64, chatID int64, weekStart time.Time) (int, error) {
	if f.weeklyWorkouts == nil {
		return 0, fmt.Errorf("unexpected WeeklyWorkouts")
	}
	return f.weeklyWorkouts(ctx, userID, chatID, weekStart)
}

func (f *fakeWorkouts) GetWorkoutCounts(ctx context.Context, chatID int64, weekStart time.Time) ([]UserWorkouts, error) {
	if f.getWorkoutCounts == nil {
		return nil, fmt.Errorf("unexpected GetWorkoutCounts")
	}
	return f.getWorkoutCounts(ctx, chatID, weekStart)
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

func (f *fakeWorkouts) GetChatStats(ctx context.Context, chatID int64, weekStart time.Time) ([]domain.UserStats, error) {
	if f.getChatStats == nil {
		return nil, fmt.Errorf("unexpected GetChatStats")
	}
	return f.getChatStats(ctx, chatID, weekStart)
}

func (f *fakeWorkouts) GetActiveChallengeVotersCount(ctx context.Context, chatID int64) (int, error) {
	if f.getVotersCount == nil {
		return 0, fmt.Errorf("unexpected GetActiveChallengeVotersCount")
	}
	return f.getVotersCount(ctx, chatID)
}

type fakeChallenges struct {
	getActiveChallenge     func(ctx context.Context) (*domain.Challenge, error)
	getActiveByChat        func(ctx context.Context, chatID int64) (*domain.Challenge, error)
	getAllActiveChallenges func(ctx context.Context) ([]domain.Challenge, error)
	getChallenge           func(ctx context.Context, challengeID int) (*domain.Challenge, error)
	hasActiveChallengeChat func(ctx context.Context, chatID int64) (bool, error)
	createChallenge        func(ctx context.Context, challenge domain.Challenge) error
	updateChallenge        func(ctx context.Context, challenge domain.Challenge) error
	deleteChallenge        func(ctx context.Context, challengeID int) error
	deactivateForChat      func(ctx context.Context, chatID int64) error
	markWeeklyDone         func(ctx context.Context, challengeID int) error
	markDailyDone          func(ctx context.Context, challengeID int) error
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

func (f *fakeChallenges) GetChallenge(ctx context.Context, challengeID int) (*domain.Challenge, error) {
	if f.getChallenge == nil {
		return nil, fmt.Errorf("unexpected GetChallenge")
	}
	return f.getChallenge(ctx, challengeID)
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

func (f *fakeChallenges) UpdateChallenge(ctx context.Context, challenge domain.Challenge) error {
	if f.updateChallenge == nil {
		return fmt.Errorf("unexpected UpdateChallenge")
	}
	return f.updateChallenge(ctx, challenge)
}

func (f *fakeChallenges) DeleteChallenge(ctx context.Context, challengeID int) error {
	if f.deleteChallenge == nil {
		return fmt.Errorf("unexpected DeleteChallenge")
	}
	return f.deleteChallenge(ctx, challengeID)
}

func (f *fakeChallenges) DeactivateChallengeForChat(ctx context.Context, chatID int64) error {
	if f.deactivateForChat == nil {
		return fmt.Errorf("unexpected DeactivateChallengeForChat")
	}
	return f.deactivateForChat(ctx, chatID)
}

func (f *fakeChallenges) MarkWeeklyCheckDone(ctx context.Context, challengeID int) error {
	if f.markWeeklyDone == nil {
		return nil // default to success in tests unless specified
	}
	return f.markWeeklyDone(ctx, challengeID)
}

func (f *fakeChallenges) MarkDailyStatsDone(ctx context.Context, challengeID int) error {
	if f.markDailyDone == nil {
		return nil // default to success in tests unless specified
	}
	return f.markDailyDone(ctx, challengeID)
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
	getVotersCount   func(ctx context.Context, chatID int64) (int, error)
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

func (f *fakeModeration) GetActiveChallengeVotersCount(ctx context.Context, chatID int64) (int, error) {
	if f.getVotersCount == nil {
		return 0, fmt.Errorf("unexpected GetActiveChallengeVotersCount")
	}
	return f.getVotersCount(ctx, chatID)
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
		weeklyWorkouts: func(ctx context.Context, userID int64, chatID int64, weekStart time.Time) (int, error) {
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
		createUser:      func(ctx context.Context, user domain.User) error { return nil },
		getAllActive:    func(ctx context.Context) ([]domain.User, error) { return nil, nil },
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
		weeklyWorkouts: func(ctx context.Context, userID int64, chatID int64, weekStart time.Time) (int, error) { return 0, nil },
	}

	challenges := &fakeChallenges{
		hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) { return true, nil },
		getActiveChallenge:     func(ctx context.Context) (*domain.Challenge, error) { return nil, errors.New("unexpected") },
		createChallenge:        func(ctx context.Context, challenge domain.Challenge) error { return nil },
		deactivateForChat:      func(ctx context.Context, chatID int64) error { return nil },
	}

	bootstrap := &fakeBootstrap{
		initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
			return nil
		},
		getBootstrap:     func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) { return nil, sql.ErrNoRows },
		setBotAdmin:      func(ctx context.Context, chatID int64, isBotAdmin bool) error { return nil },
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
			return nil
		},
		countHeartReactions: func(ctx context.Context, chatID int64) (int, error) { return 0, nil },
		markStarted:         func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
		updateConfig: func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
			return nil
		},
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
		createUser:      func(ctx context.Context, user domain.User) error { return nil },
		readUser:        func(ctx context.Context, telegramID int64) (*domain.User, error) { return &domain.User{}, nil },
		getAllActive:    func(ctx context.Context) ([]domain.User, error) { return nil, nil },
		batchDeactivate: func(ctx context.Context, userIDs []int64) error { return nil },
	}

	bootstrap := &fakeBootstrap{
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
			return nil
		},
		getBootstrap:     func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) { return nil, sql.ErrNoRows },
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
		initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
			return nil
		},
		setBotAdmin:         func(ctx context.Context, chatID int64, isBotAdmin bool) error { return nil },
		countHeartReactions: func(ctx context.Context, chatID int64) (int, error) { return 0, nil },
		markStarted:         func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
		updateConfig: func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
			return nil
		},
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
		markStarted:         func(ctx context.Context, chatID int64) (bool, error) { return true, nil },
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
			return nil
		},
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
		initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
			return nil
		},
		setBotAdmin: func(ctx context.Context, chatID int64, isBotAdmin bool) error { return nil },
	}

	created := false
	challenges := &fakeChallenges{
		createChallenge: func(ctx context.Context, challenge domain.Challenge) error {
			created = true
			return nil
		},
		getActiveChallenge:     func(ctx context.Context) (*domain.Challenge, error) { return nil, errors.New("unexpected") },
		deactivateForChat:      func(ctx context.Context, chatID int64) error { return nil },
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
		getActiveChallenge:     func(ctx context.Context) (*domain.Challenge, error) { return challenge, nil },
		createChallenge:        func(ctx context.Context, challenge domain.Challenge) error { return nil },
		deactivateForChat:      func(ctx context.Context, chatID int64) error { return nil },
		hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
	}

	users := &fakeUsers{
		createUser: func(ctx context.Context, user domain.User) error { return nil },
		readUser:   func(ctx context.Context, telegramID int64) (*domain.User, error) { return &domain.User{}, nil },
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
		createWorkout:   func(ctx context.Context, workout domain.Workout) error { return nil },
		getWorkoutCounts: func(ctx context.Context, chatID int64, weekStart time.Time) ([]UserWorkouts, error) {
			expectedWeekStart := getWeekStart(time.Now())
			if weekStart.Unix() != expectedWeekStart.Unix() {
				return nil, fmt.Errorf("unexpected week start: got %v, want %v", weekStart, expectedWeekStart)
			}
			return []UserWorkouts{
				{UserInfo: UserInfo{TelegramID: 10, Username: "a"}, Count: 2},
				{UserInfo: UserInfo{TelegramID: 11, Username: "b"}, Count: 3},
			}, nil
		},
	}

	svc := New(Deps{
		Users:     users,
		Sessions:  &fakeSessions{},
		Workouts:  workouts,
		Challenge: challenges,
		Bootstrap: &fakeBootstrap{
			initBootstrap: func(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
				return nil
			},
			getBootstrap:     func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) { return nil, sql.ErrNoRows },
			setBotAdmin:      func(ctx context.Context, chatID int64, isBotAdmin bool) error { return nil },
			upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error { return nil },
			setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
				return nil
			},
			countHeartReactions: func(ctx context.Context, chatID int64) (int, error) { return 0, nil },
			markStarted:         func(ctx context.Context, chatID int64) (bool, error) { return false, nil },
			updateConfig: func(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
				return nil
			},
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
		Users:      &fakeUsers{createUser: func(ctx context.Context, user domain.User) error { return nil }},
		Bootstrap:  bootstrap,
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

func TestChallengeService_GetStats(t *testing.T) {
	ctx := context.Background()

	startedAt := time.Now().Add(-10 * 24 * time.Hour) // 10 days ago (week 1)
	challenge := &domain.Challenge{ChatID: 1, StartedAt: startedAt}

	challenges := &fakeChallenges{
		getActiveByChat: func(ctx context.Context, chatID int64) (*domain.Challenge, error) {
			return challenge, nil
		},
	}

	workouts := &fakeWorkouts{
		getChatStats: func(ctx context.Context, chatID int64, weekStart time.Time) ([]domain.UserStats, error) {
			expectedWeekStart := getWeekStart(time.Now())
			if weekStart.Unix() != expectedWeekStart.Unix() {
				return nil, fmt.Errorf("unexpected week start: got %v, want %v", weekStart, expectedWeekStart)
			}

			return []domain.UserStats{
				{Username: "user1", WeeklyCount: 3, TotalCount: 5, IsActive: true},
				{Username: "user2", WeeklyCount: 1, TotalCount: 2, IsActive: false},
			}, nil
		},
	}

	svc := New(Deps{
		Challenge: challenges,
		Workouts:  workouts,
	})

	stats, err := svc.GetStats(ctx, 1)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}

	if stats[0].Username != "user1" || !stats[0].IsActive {
		t.Errorf("unexpected stat for user1")
	}

}

func TestUserService_CRUD(t *testing.T) {
	ctx := context.Background()

	users := &fakeUsers{
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
			return &domain.User{TelegramID: telegramID, Username: "test"}, nil
		},
		updateUser: func(ctx context.Context, user domain.User) error {
			return nil
		},
		deleteUser: func(ctx context.Context, telegramID int64) error {
			return nil
		},
		getAllActive: func(ctx context.Context) ([]domain.User, error) {
			return []domain.User{{TelegramID: 1, Username: "test"}}, nil
		},
		getUserIDByUsername: func(ctx context.Context, username string) (int64, error) {
			return 1, nil
		},
	}

	svc := New(Deps{Users: users})

	_, err := svc.ReadUser(ctx, 1)
	if err != nil {
		t.Errorf("ReadUser error: %v", err)
	}

	err = svc.UpdateUser(ctx, domain.User{TelegramID: 1})
	if err != nil {
		t.Errorf("UpdateUser error: %v", err)
	}

	err = svc.DeleteUser(ctx, 1)
	if err != nil {
		t.Errorf("DeleteUser error: %v", err)
	}

	_, err = svc.GetAllActiveUsers(ctx)
	if err != nil {
		t.Errorf("GetAllActiveUsers error: %v", err)
	}

	_, err = svc.GetUserIDByUsername(ctx, "test")
	if err != nil {
		t.Errorf("GetUserIDByUsername error: %v", err)
	}
}

func TestUserService_IsActiveUser_Coverage(t *testing.T) {
	ctx := context.Background()

	users := &fakeUsers{
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
			if telegramID == 1 {
				return &domain.User{TelegramID: 1, IsActive: true}, nil
			}
			if telegramID == 2 {
				return &domain.User{TelegramID: 2, IsActive: false}, nil
			}
			if telegramID == 3 {
				return nil, sql.ErrNoRows
			}
			return nil, errors.New("db error")
		},
	}

	svc := New(Deps{Users: users})

	active, err := svc.Users.IsActiveUser(ctx, 1)
	if err != nil || !active {
		t.Errorf("Expected active user 1")
	}

	active, err = svc.Users.IsActiveUser(ctx, 2)
	if err != nil || active {
		t.Errorf("Expected inactive user 2")
	}

	active, err = svc.Users.IsActiveUser(ctx, 3)
	if err != nil || active {
		t.Errorf("Expected inactive user 3 (not found)")
	}

	_, err = svc.Users.IsActiveUser(ctx, 4)
	if err == nil {
		t.Errorf("Expected error for user 4")
	}
}

func TestChallengeService_CRUD(t *testing.T) {
	ctx := context.Background()

	challenges := &fakeChallenges{
		getChallenge: func(ctx context.Context, challengeID int) (*domain.Challenge, error) {
			return &domain.Challenge{ChallengeID: challengeID}, nil
		},
		updateChallenge: func(ctx context.Context, challenge domain.Challenge) error {
			return nil
		},
		deleteChallenge: func(ctx context.Context, challengeID int) error {
			return nil
		},
		markWeeklyDone: func(ctx context.Context, challengeID int) error {
			return nil
		},
		markDailyDone: func(ctx context.Context, challengeID int) error {
			return nil
		},
	}

	users := &fakeUsers{
		getUserIDByUsername: func(ctx context.Context, username string) (int64, error) {
			return 1, nil
		},
	}

	workouts := &fakeWorkouts{
		addWorkouts: func(ctx context.Context, userID int64, chatID int64, amount int) (int, error) {
			return amount, nil
		},
	}

	svc := New(Deps{Challenge: challenges, Users: users, Workouts: workouts})

	_, err := svc.GetChallenge(ctx, 1)
	if err != nil {
		t.Errorf("GetChallenge error: %v", err)
	}

	err = svc.UpdateChallenge(ctx, domain.Challenge{ChallengeID: 1})
	if err != nil {
		t.Errorf("UpdateChallenge error: %v", err)
	}

	err = svc.DeleteChallenge(ctx, 1)
	if err != nil {
		t.Errorf("DeleteChallenge error: %v", err)
	}

	err = svc.MarkWeeklyCheckDone(ctx, 1)
	if err != nil {
		t.Errorf("MarkWeeklyCheckDone error: %v", err)
	}

	err = svc.MarkDailyStatsDone(ctx, 1)
	if err != nil {
		t.Errorf("MarkDailyStatsDone error: %v", err)
	}

	added, err := svc.AddWorkoutDirect(ctx, 1, "test", 2)
	if err != nil || added != 2 {
		t.Errorf("AddWorkoutDirect error: %v, added: %d", err, added)
	}
}

func TestModerationService_Comprehensive(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(*fakeWorkouts, *fakeSessions, *fakeVotes, *fakeUsers)
		execute    func(*Service) error
		wantErrStr string
	}{
		{
			name: "DisputeWorkout_Success",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getWorkoutByMsg = func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
					return &domain.Workout{ID: 1, UserID: 1, WorkoutDate: time.Now()}, nil
				}
				w.cancelWorkout = func(ctx context.Context, chatID int64, messageID int, cancelledBy int64) (int64, error) {
					return 1, nil
				}
			},
			execute: func(svc *Service) error {
				_, _, err := svc.DisputeWorkout(ctx, 1, 1, 2)
				return err
			},
			wantErrStr: "",
		},
		{
			name: "DisputeWorkout_SelfCancel",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getWorkoutByMsg = func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
					return &domain.Workout{ID: 1, UserID: 2, WorkoutDate: time.Now()}, nil // Same as disputer
				}
			},
			execute: func(svc *Service) error {
				_, _, err := svc.DisputeWorkout(ctx, 1, 1, 2)
				return err
			},
			wantErrStr: "не можна скасовувати власне тренування",
		},
		{
			name: "DisputeWorkout_Timeout",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getWorkoutByMsg = func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
					return &domain.Workout{ID: 1, UserID: 1, WorkoutDate: time.Now().Add(-15 * time.Minute)}, nil
				}
			},
			execute: func(svc *Service) error {
				_, _, err := svc.DisputeWorkout(ctx, 1, 1, 2)
				return err
			},
			wantErrStr: "час для скасування вичерпано (10 хв)",
		},
		{
			name: "DisputeSession_Success",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getWorkoutByMsg = func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
					return nil, sql.ErrNoRows
				}
				s.getSessionByMessage = func(ctx context.Context, chatID int64, messageID int) (*domain.Session, error) {
					return &domain.Session{UserID: 1, StartedAt: time.Now()}, nil
				}
				s.deleteSession = func(ctx context.Context, chatID int64, messageID int) error { return nil }
			},
			execute: func(svc *Service) error {
				_, isSession, err := svc.DisputeWorkout(ctx, 1, 1, 2)
				if !isSession {
					return fmt.Errorf("expected isSession to be true")
				}
				return err
			},
			wantErrStr: "",
		},
		{
			name: "ReinstateWorkout_Success",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getWorkoutByMsg = func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
					cb := int64(2)
					return &domain.Workout{ID: 1, UserID: 1, WorkoutDate: time.Now(), CancelledBy: &cb}, nil
				}
				w.reinstateWorkout = func(ctx context.Context, chatID int64, messageID int, reinstatedBy int64) error {
					return nil
				}
			},
			execute: func(svc *Service) error {
				return svc.ReinstateWorkout(ctx, 1, 1, 2)
			},
			wantErrStr: "",
		},
		{
			name: "HandlePollUpdate_SubtractSuccess",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				v.getVoteByPollID = func(ctx context.Context, pollID string) (*domain.Vote, error) {
					return &domain.Vote{ID: 1, Type: "subtract", TargetUserID: 1, ChatID: 1, Amount: 2, IsCompleted: false}, nil
				}
				v.completeVote = func(ctx context.Context, pollID string, success bool) error { return nil }
				w.subtractWorkouts = func(ctx context.Context, userID int64, chatID int64, amount int) (int, error) { return amount, nil }
			},
			execute: func(svc *Service) error {
				_, err := svc.HandlePollUpdate(ctx, "poll1", true)
				return err
			},
			wantErrStr: "",
		},
		{
			name: "HandlePollUpdate_AlreadyCompleted",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				v.getVoteByPollID = func(ctx context.Context, pollID string) (*domain.Vote, error) {
					return &domain.Vote{ID: 1, Type: "subtract", IsCompleted: true}, nil
				}
			},
			execute: func(svc *Service) error {
				_, err := svc.HandlePollUpdate(ctx, "poll1", true)
				return err
			},
			wantErrStr: "",
		},
		{
			name: "InitiateSubtract_Success",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				u.getUserIDByUsername = func(ctx context.Context, username string) (int64, error) {
					return 1, nil
				}
				v.createVote = func(ctx context.Context, vote domain.Vote) error {
					return nil
				}
			},
			execute: func(svc *Service) error {
				return svc.InitiateSubtract(ctx, 1, 2, "test", 1, "poll1")
			},
			wantErrStr: "",
		},
		{
			name: "InitiateSubtract_SelfVote",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				u.getUserIDByUsername = func(ctx context.Context, username string) (int64, error) {
					return 1, nil
				}
			},
			execute: func(svc *Service) error {
				return svc.InitiateSubtract(ctx, 1, 1, "test", 1, "poll1")
			},
			wantErrStr: "не можна голосувати проти себе",
		},
		{
			name: "InitiateAdd_Success",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				u.getUserIDByUsername = func(ctx context.Context, username string) (int64, error) {
					return 1, nil
				}
				v.createVote = func(ctx context.Context, vote domain.Vote) error {
					return nil
				}
			},
			execute: func(svc *Service) error {
				return svc.InitiateAdd(ctx, 1, 2, "test", 1, "poll2")
			},
			wantErrStr: "",
		},
		{
			name: "GetWorkoutByMessage",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getWorkoutByMsg = func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
					return &domain.Workout{ID: 1}, nil
				}
			},
			execute: func(svc *Service) error {
				_, err := svc.GetWorkoutByMessage(ctx, 1, 1)
				return err
			},
			wantErrStr: "",
		},
		{
			name: "GetActiveChallengeVotersCount",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getVotersCount = func(ctx context.Context, chatID int64) (int, error) {
					return 5, nil
				}
			},
			execute: func(svc *Service) error {
				_, err := svc.GetActiveChallengeVotersCount(ctx, 1)
				return err
			},
			wantErrStr: "",
		},
		{
			name: "ReinstateWorkout_NotCancelledByYou",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getWorkoutByMsg = func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
					cb := int64(3) // cancelled by someone else
					return &domain.Workout{ID: 1, UserID: 1, WorkoutDate: time.Now(), CancelledBy: &cb}, nil
				}
			},
			execute: func(svc *Service) error {
				return svc.ReinstateWorkout(ctx, 1, 1, 2)
			},
			wantErrStr: "лише той, хто скасував, може повернути тренування",
		},
		{
			name: "ReinstateWorkout_Timeout",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				w.getWorkoutByMsg = func(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) {
					cb := int64(2)
					return &domain.Workout{ID: 1, UserID: 1, WorkoutDate: time.Now().Add(-15 * time.Minute), CancelledBy: &cb}, nil
				}
			},
			execute: func(svc *Service) error {
				return svc.ReinstateWorkout(ctx, 1, 1, 2)
			},
			wantErrStr: "час для повернення вичерпано",
		},
		{
			name: "HandlePollUpdate_AddSuccess",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				v.getVoteByPollID = func(ctx context.Context, pollID string) (*domain.Vote, error) {
					return &domain.Vote{ID: 1, Type: "add", TargetUserID: 1, ChatID: 1, Amount: 2, IsCompleted: false}, nil
				}
				v.completeVote = func(ctx context.Context, pollID string, success bool) error { return nil }
				w.addWorkouts = func(ctx context.Context, userID int64, chatID int64, amount int) (int, error) { return amount, nil }
			},
			execute: func(svc *Service) error {
				_, err := svc.HandlePollUpdate(ctx, "poll_add", true)
				return err
			},
			wantErrStr: "",
		},
		{
			name: "HandlePollUpdate_FailureDoesNothing",
			setup: func(w *fakeWorkouts, s *fakeSessions, v *fakeVotes, u *fakeUsers) {
				v.getVoteByPollID = func(ctx context.Context, pollID string) (*domain.Vote, error) {
					return &domain.Vote{ID: 1, Type: "add", TargetUserID: 1, ChatID: 1, Amount: 2, IsCompleted: false}, nil
				}
				v.completeVote = func(ctx context.Context, pollID string, success bool) error { return nil }
			},
			execute: func(svc *Service) error {
				_, err := svc.HandlePollUpdate(ctx, "poll_fail", false)
				return err
			},
			wantErrStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &fakeWorkouts{}
			s := &fakeSessions{}
			v := &fakeVotes{}
			u := &fakeUsers{}
			tt.setup(w, s, v, u)
			svc := New(Deps{Workouts: w, Sessions: s, Votes: v, Users: u})

			err := tt.execute(svc)
			if tt.wantErrStr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.wantErrStr {
					t.Errorf("expected error '%s', got %v", tt.wantErrStr, err)
				}
			}
		})
	}
}

func TestBootstrapService_GetConfig(t *testing.T) {
	ctx := context.Background()

	bootstrap := &fakeBootstrap{
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
			if chatID == 1 {
				return &domain.ChallengeBootstrap{ChatID: 1}, nil
			}
			return nil, sql.ErrNoRows
		},
	}

	challenges := &fakeChallenges{
		getActiveByChat: func(ctx context.Context, chatID int64) (*domain.Challenge, error) {
			return nil, sql.ErrNoRows
		},
	}

	svc := New(Deps{Bootstrap: bootstrap, Challenge: challenges})

	_, err := svc.GetChallengeConfig(ctx, 1)
	if err != nil {
		t.Errorf("GetChallengeConfig error: %v", err)
	}

	_, err = svc.GetChallengeConfig(ctx, 2)
	if err == nil {
		t.Errorf("Expected error for non-existent config")
	}
}

func TestBootstrapService_ProcessReactionUpdate(t *testing.T) {
	ctx := context.Background()

	users := &fakeUsers{
		getUserIDByUsername: func(ctx context.Context, username string) (int64, error) {
			return 1, nil
		},
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
			return nil, sql.ErrNoRows // simulate new user to hit CreateUser
		},
		createUser: func(ctx context.Context, user domain.User) error {
			return nil
		},
	}

	bootstrap := &fakeBootstrap{
		setReactions: func(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
			return nil
		},
		getBootstrap: func(ctx context.Context, chatID int64) (*domain.ChallengeBootstrap, error) {
			return &domain.ChallengeBootstrap{WelcomeMessageID: 1, ExpectedReactions: 1, IsStarted: false, IsBotAdmin: true}, nil
		},
		countHeartReactions: func(ctx context.Context, chatID int64) (int, error) {
			return 1, nil
		},
		markStarted: func(ctx context.Context, chatID int64) (bool, error) {
			return true, nil
		},
		upsertChatMember: func(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error {
			return nil
		},
	}

	mod := &fakeModeration{
		cancelCounted: func(ctx context.Context, chatID int64, messageID int) (bool, error) {
			return true, nil
		},
	}

	challenges := &fakeChallenges{
		createChallenge: func(ctx context.Context, challenge domain.Challenge) error {
			return nil
		},
	}

	svc := New(Deps{Users: users, Bootstrap: bootstrap, Moderation: mod, Challenge: challenges})

	// Test Heart reaction -> Challenge Started
	started, cancelled, err := svc.ProcessReactionUpdate(ctx, 1, 1, 1, "test", []string{"❤️"})
	if err != nil || !started || cancelled {
		t.Errorf("ProcessReactionUpdate Heart error: %v, started: %v, cancelled: %v", err, started, cancelled)
	}

	// Test ThumbsDown reaction -> Cancel Workout
	started, cancelled, err = svc.ProcessReactionUpdate(ctx, 1, 2, 1, "test", []string{"👎"})
	if err != nil || started || !cancelled {
		t.Errorf("ProcessReactionUpdate ThumbsDown error: %v, started: %v, cancelled: %v", err, started, cancelled)
	}
}

func TestCircleService_HandleCircle(t *testing.T) {
	ctx := context.Background()

	users := &fakeUsers{
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
			if telegramID == 1 {
				return &domain.User{IsActive: true}, nil
			}
			return nil, sql.ErrNoRows
		},
	}

	challenges := &fakeChallenges{
		hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) {
			return chatID == 1, nil
		},
	}

	sessions := &fakeSessions{
		hasTrainedToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) {
			return false, nil
		},
		getSession: func(ctx context.Context, userID int64, chatID int64) (*domain.Session, error) {
			return nil, sql.ErrNoRows // no active session
		},
		startSession: func(ctx context.Context, userID int64, chatID int64, messageID int) error {
			return nil
		},
	}

	workouts := &fakeWorkouts{
		hasWorkoutToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) {
			return false, nil
		},
	}

	svc := New(Deps{
		Users:     users,
		Challenge: challenges,
		Sessions:  sessions,
		Workouts:  workouts,
		Rules: Rules{
			MinCircleDurationSeconds: 10,
			SessionGap:               5 * time.Minute,
		},
	})

	// 1. Success - short circle starts session
	err := svc.HandleCircle(ctx, 1, 15, 1, 100)
	if err != nil {
		t.Errorf("HandleCircle error: %v", err)
	}

	// 2. Too short
	err = svc.HandleCircle(ctx, 1, 5, 1, 101)
	if err != nil {
		t.Errorf("Expected no error for short circle, got: %v", err)
	}

	// 3. User not active
	err = svc.HandleCircle(ctx, 2, 15, 1, 102)
	if err != nil {
		t.Errorf("Expected no error for inactive user, got: %v", err)
	}

	// 4. Chat has no active challenge
	err = svc.HandleCircle(ctx, 1, 15, 2, 103)
	if err != nil {
		t.Errorf("Expected no error for inactive chat, got: %v", err)
	}
}

func TestCircleService_CompleteWorkout(t *testing.T) {
	ctx := context.Background()

	users := &fakeUsers{
		readUser: func(ctx context.Context, telegramID int64) (*domain.User, error) {
			return &domain.User{IsActive: true}, nil
		},
	}

	challenges := &fakeChallenges{
		hasActiveChallengeChat: func(ctx context.Context, chatID int64) (bool, error) {
			return true, nil
		},
	}

	sessions := &fakeSessions{
		hasTrainedToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) {
			return false, nil
		},
		getSession: func(ctx context.Context, userID int64, chatID int64) (*domain.Session, error) {
			return &domain.Session{StartedAt: time.Now().Add(-10 * time.Minute), LastVideoAt: time.Now()}, nil
		},
		addLatestSession: func(ctx context.Context, userID int64, chatID int64) error {
			return nil
		},
		startSession: func(ctx context.Context, userID int64, chatID int64, messageID int) error {
			return nil
		},
		deleteSession: func(ctx context.Context, chatID int64, messageID int) error {
			return nil
		},
	}

	workouts := &fakeWorkouts{
		createWorkout: func(ctx context.Context, workout domain.Workout) error {
			return nil
		},
		hasWorkoutToday: func(ctx context.Context, userID int64, chatID int64) (bool, error) {
			return false, nil
		},
	}

	svc := New(Deps{
		Users:     users,
		Challenge: challenges,
		Sessions:  sessions,
		Workouts:  workouts,
		Rules: Rules{
			MinCircleDurationSeconds: 10,
			SessionGap:               5 * time.Minute,
		},
	})

	err := svc.HandleCircle(ctx, 1, 15, 1, 100)
	if err != nil {
		t.Errorf("HandleCircle error: %v", err)
	}
}

func TestService_GetRules(t *testing.T) {
	svc := New(Deps{})
	rules := svc.GetRules()
	if rules.MinCircleDurationSeconds == 0 {
		t.Errorf("Expected initialized rules")
	}
}
