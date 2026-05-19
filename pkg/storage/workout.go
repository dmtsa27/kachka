package storage

import (
	"context"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/service"
)

func (s *Storage) CreateWorkout(ctx context.Context, workout Workout) error {
	query := `INSERT INTO workouts (user_id, workout_date, chat_id, completion_message_id)
	VALUES ($1, $2, $3, $4)
	
	`
	_, err := s.db.ExecContext(ctx, query, workout.UserID, workout.WorkoutDate, workout.ChatID, workout.MessageID)

	return err
}

func (s *Storage) WeeklyWorkouts(ctx context.Context, userID int64, weekStart time.Time) (int, error) {
	query := `
        SELECT COUNT(*) 
        FROM workouts 
        WHERE user_id = $1 
          AND workout_date >= $2
          AND workout_date < $2 + INTERVAL '7 days'`

	var count int

	err := s.db.QueryRowContext(ctx, query, userID, weekStart).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Storage) GetWorkoutCounts(ctx context.Context, weekStart time.Time) ([]service.UserWorkouts, error) {
	query := `
        SELECT u.telegram_id, u.username, COUNT(w.id) as workout_count
        FROM users u
        LEFT JOIN workouts w ON u.telegram_id = w.user_id 
            AND w.workout_date >= $1 
            AND w.workout_date < $1 + INTERVAL '7 days'
        WHERE u.is_active = true
        GROUP BY u.telegram_id, u.username`

	rows, err := s.db.QueryContext(ctx, query, weekStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []service.UserWorkouts
	for rows.Next() {
		var uw service.UserWorkouts
		if err := rows.Scan(&uw.TelegramID, &uw.Username, &uw.Count); err != nil {
			return nil, err
		}
		counts = append(counts, uw)
	}

	return counts, rows.Err()
}

func (s *Storage) HasWorkoutToday(ctx context.Context, userID int64, chatID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM workouts 
			WHERE user_id = $1 AND chat_id = $2 AND CAST(workout_date AS DATE) = CURRENT_DATE
		)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userID, chatID).Scan(&exists)
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

func (s *Storage) DeleteWorkoutByMessageToday(ctx context.Context, chatID int64, messageID int) (bool, error) {
	query := `
		DELETE FROM workouts
		WHERE chat_id = $1 AND completion_message_id = $2 AND CAST(workout_date AS DATE) = CURRENT_DATE
	`
	result, err := s.db.ExecContext(ctx, query, chatID, messageID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
