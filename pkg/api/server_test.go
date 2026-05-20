package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
	"github.com/dmtsa27/kachka.git/pkg/service"
)

type mockUsers struct {
	readUser func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUsers) CreateUser(ctx context.Context, user domain.User) error { return nil }
func (m *mockUsers) ReadUser(ctx context.Context, id int64) (*domain.User, error) {
	return m.readUser(ctx, id)
}
func (m *mockUsers) UpdateUser(ctx context.Context, user domain.User) error { return nil }
func (m *mockUsers) DeleteUser(ctx context.Context, id int64) error         { return nil }
func (m *mockUsers) GetAllActiveUsers(ctx context.Context) ([]domain.User, error) {
	return []domain.User{{TelegramID: 1, Username: "test"}}, nil
}
func (m *mockUsers) BatchDeactivateUsers(ctx context.Context, ids []int64) error        { return nil }
func (m *mockUsers) GetUserIDByUsername(ctx context.Context, username string) (int64, error) { return 1, nil }

func TestHandlePing(t *testing.T) {
	svc := service.New(service.Deps{})
	server := NewServer(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"message":"pong"}` {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestHandleGetUser(t *testing.T) {
	m := &mockUsers{
		readUser: func(ctx context.Context, id int64) (*domain.User, error) {
			return &domain.User{TelegramID: id, Username: "test"}, nil
		},
	}
	svc := service.New(service.Deps{Users: m})
	server := NewServer(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/users/123", nil)
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleCreateUser(t *testing.T) {
	svc := service.New(service.Deps{Users: &mockUsers{}})
	server := NewServer(svc)

	user := domain.User{TelegramID: 123, Username: "test"}
	body, _ := json.Marshal(user)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/admin/users", bytes.NewBuffer(body))
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestHandleUpdateUser(t *testing.T) {
	svc := service.New(service.Deps{Users: &mockUsers{}})
	server := NewServer(svc)

	user := domain.User{Username: "updated"}
	body, _ := json.Marshal(user)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/admin/users/123", bytes.NewBuffer(body))
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleDeleteUser(t *testing.T) {
	svc := service.New(service.Deps{Users: &mockUsers{}})
	server := NewServer(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/admin/users/123", nil)
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

type mockChallenges struct {
	activeChallenges func(ctx context.Context) ([]domain.Challenge, error)
	getChallenge     func(ctx context.Context, id int) (*domain.Challenge, error)
}

func (m *mockChallenges) CreateChallenge(ctx context.Context, ch domain.Challenge) error { return nil }
func (m *mockChallenges) GetChallenge(ctx context.Context, id int) (*domain.Challenge, error) {
	return m.getChallenge(ctx, id)
}
func (m *mockChallenges) UpdateChallenge(ctx context.Context, ch domain.Challenge) error { return nil }
func (m *mockChallenges) DeleteChallenge(ctx context.Context, id int) error            { return nil }
func (m *mockChallenges) ActiveChallenges(ctx context.Context) ([]domain.Challenge, error) {
	return m.activeChallenges(ctx)
}
func (m *mockChallenges) HasActiveChallengeInChat(ctx context.Context, chatID int64) (bool, error) {
	return true, nil
}
func (m *mockChallenges) DeactivateChallengeForChat(ctx context.Context, chatID int64) error { return nil }
func (m *mockChallenges) GetActiveChallenge(ctx context.Context) (*domain.Challenge, error) { return nil, nil }
func (m *mockChallenges) GetActiveChallengeByChat(ctx context.Context, chatID int64) (*domain.Challenge, error) { return nil, nil }
func (m *mockChallenges) GetAllActiveChallenges(ctx context.Context) ([]domain.Challenge, error) { return nil, nil }
func (m *mockChallenges) MarkWeeklyCheckDone(ctx context.Context, id int) error            { return nil }
func (m *mockChallenges) MarkDailyStatsDone(ctx context.Context, id int) error             { return nil }
func (m *mockChallenges) SetWeekRules(ctx context.Context, id int, rules string) error      { return nil }

type mockWorkouts struct {
	addWorkouts func(ctx context.Context, userID int64, chatID int64, amount int) (int, error)
}

func (m *mockWorkouts) CreateWorkout(ctx context.Context, w domain.Workout) error { return nil }
func (m *mockWorkouts) WeeklyWorkouts(ctx context.Context, userID int64, chatID int64, weekStart time.Time) (int, error) { return 0, nil }
func (m *mockWorkouts) GetWorkoutCounts(ctx context.Context, chatID int64, weekStart time.Time) ([]service.UserWorkouts, error) { return nil, nil }
func (m *mockWorkouts) HasWorkoutToday(ctx context.Context, userID int64, chatID int64) (bool, error) { return false, nil }
func (m *mockWorkouts) CancelWorkout(ctx context.Context, chatID int64, messageID int, cancelledBy int64) (int64, error) { return 0, nil }
func (m *mockWorkouts) ReinstateWorkout(ctx context.Context, chatID int64, messageID int, reinstatedBy int64) error { return nil }
func (m *mockWorkouts) GetWorkoutByMessage(ctx context.Context, chatID int64, messageID int) (*domain.Workout, error) { return nil, nil }
func (m *mockWorkouts) SubtractWorkouts(ctx context.Context, userID int64, chatID int64, amount int) (int, error) { return 0, nil }
func (m *mockWorkouts) AddWorkouts(ctx context.Context, userID int64, chatID int64, amount int) (int, error) {
	return m.addWorkouts(ctx, userID, chatID, amount)
}
func (m *mockWorkouts) GetChatStats(ctx context.Context, chatID int64, weekStart time.Time) ([]domain.UserStats, error) { return nil, nil }
func (m *mockWorkouts) GetActiveChallengeVotersCount(ctx context.Context, chatID int64) (int, error) { return 0, nil }

func TestHandleGetChallenges(t *testing.T) {
	m := &mockChallenges{
		activeChallenges: func(ctx context.Context) ([]domain.Challenge, error) {
			return []domain.Challenge{{ChallengeID: 1}}, nil
		},
	}
	svc := service.New(service.Deps{Challenge: m})
	server := NewServer(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/challenges", nil)
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetChallenge(t *testing.T) {
	m := &mockChallenges{
		getChallenge: func(ctx context.Context, id int) (*domain.Challenge, error) {
			return &domain.Challenge{ChallengeID: id}, nil
		},
	}
	svc := service.New(service.Deps{Challenge: m})
	server := NewServer(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/challenges/1", nil)
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleAddWorkout(t *testing.T) {
	svc := service.New(service.Deps{
		Users:    &mockUsers{},
		Workouts: &mockWorkouts{addWorkouts: func(ctx context.Context, userID int64, chatID int64, amount int) (int, error) { return amount, nil }},
	})
	server := NewServer(svc)

	reqBody := AddWorkoutRequest{Username: "test", ChatID: 1, Amount: 1}
	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/admin/workouts", bytes.NewBuffer(body))
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}


