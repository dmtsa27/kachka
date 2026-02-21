package storage

import (
	"context"
	"time"
)

type Workout struct {
	WorkoutDate time.Time
	ID          int
	UserID      int64
}

func (s *Storage) CreateWorkout(ctx context.Context, workout Workout) error {
	query := `INSERT INTO workouts (user_id, workout_date)
	VALUES ($1, $2)
	
	`
	_, err := s.db.ExecContext(ctx, query, workout.UserID, workout.WorkoutDate)

	return err
}

func (s *Storage) WeeklyWorkouts(ctx context.Context, userID int64) (int, error) {
	query := `
        SELECT COUNT(*) 
        FROM workouts 
        WHERE user_id = $1 
          AND workout_date >= NOW() - INTERVAL '7 days'`

	var count int

	err := s.db.QueryRowContext(ctx, query, userID).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Storage) HasWorkoutToday(ctx context.Context, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM workouts 
			WHERE user_id = $1 AND CAST(workout_date AS DATE) = CURRENT_DATE
		)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Storage) RemoveWorkout(ctx context.Context, workoutID int) error {
	query := `
		DELETE FROM workouts WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, query, workoutID)

	if err != nil {
		return err
	}
	return nil
}
