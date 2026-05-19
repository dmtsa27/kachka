package service

import (
	"context"
	"testing"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

func TestWeeklyCheck_NPlusOneFixed(t *testing.T) {
	ctx := context.Background()
	userCount := 10
	queryCount := 0

	users := &fakeUsers{
		batchDeactivate: func(ctx context.Context, userIDs []int64) error { return nil },
	}

	workouts := &fakeWorkouts{
		getWorkoutCounts: func(ctx context.Context, weekStart time.Time) ([]UserWorkouts, error) {
			queryCount++ // This should be 1 now
			res := make([]UserWorkouts, userCount)
			for i := 0; i < userCount; i++ {
				res[i] = UserWorkouts{
					UserInfo: UserInfo{TelegramID: int64(i), Username: "user"},
					Count:    5,
				}
			}
			return res, nil
		},
	}

	challenge := domain.Challenge{ChatID: 1, DaysPerWeek: 3, StartedAt: time.Now().Add(-8 * 24 * time.Hour)}

	svc := &challengeService{
		users:    users,
		workouts: workouts,
	}

	_, err := svc.WeeklyCheck(ctx, challenge)
	if err != nil {
		t.Fatalf("WeeklyCheck error: %v", err)
	}

	if queryCount != 1 {
		t.Errorf("Expected exactly 1 query, but got %d", queryCount)
	} else {
		t.Logf("N+1 fixed: 1 query for %d users", userCount)
	}
}

func TestActiveChallenges_MultipleChallengesFixed(t *testing.T) {
	ctx := context.Background()

	challenges := &fakeChallenges{
		getAllActiveChallenges: func(ctx context.Context) ([]domain.Challenge, error) {
			return []domain.Challenge{
				{ChatID: 100, StartedAt: time.Now()},
				{ChatID: 200, StartedAt: time.Now()},
			}, nil
		},
	}

	svc := &challengeService{
		challenge: challenges,
	}

	res, err := svc.ActiveChallenges(ctx)
	if err != nil {
		t.Fatalf("ActiveChallenges error: %v", err)
	}

	if len(res) != 2 {
		t.Errorf("Expected 2 challenges, got %d", len(res))
	}

	if res[0].ChatID != 100 || res[1].ChatID != 200 {
		t.Errorf("Challenge chat IDs mismatch")
	}
}
