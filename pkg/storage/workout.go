package storage

import (
	"context"
	"time"
)

type Workout struct {
	StartTime time.Time
	EndTime   time.Time
	WorkoutId int
	ID        int
	UserID    int64
	Duration  int
}

func (s *Storage) CreateWorkout(ctx context.Context, workout Workout) error {
	query := `INSERT INTO workouts (ide)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id) DO NOTHING
	`
	_, err := s.db.ExecContext(ctx, query, workout.UserID)

	return err
}
