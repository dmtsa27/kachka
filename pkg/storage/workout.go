package storage

import (
	"context"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/service"
)

func (s *Storage) CreateWorkout(ctx context.Context, workout Workout) error {
	query := `INSERT INTO workouts (user_id, workout_date, chat_id, completion_message_id)
	VALUES ($1, $2, $3, $4)`
	_, err := s.db.ExecContext(ctx, query, workout.UserID, workout.WorkoutDate, workout.ChatID, workout.MessageID)
	return err
}

func (s *Storage) WeeklyWorkouts(ctx context.Context, userID int64, weekStart time.Time) (int, error) {
	query := `
        SELECT COUNT(*) 
        FROM workouts 
        WHERE user_id = $1 
          AND workout_date >= $2
          AND workout_date < $2 + INTERVAL '7 days'
          AND is_cancelled = false`

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
            AND w.is_cancelled = false
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
			WHERE user_id = $1 AND chat_id = $2 AND CAST(workout_date AS DATE) = CURRENT_DATE AND is_cancelled = false
		)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userID, chatID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Storage) CancelWorkout(ctx context.Context, chatID int64, messageID int, cancelledBy int64) (int64, error) {
	var targetUserID int64
	err := s.db.QueryRowContext(ctx, `
		UPDATE workouts 
		SET is_cancelled = true, cancelled_by = $3, cancelled_at = NOW() 
		WHERE chat_id = $1 AND completion_message_id = $2 AND is_cancelled = false
		RETURNING user_id`,
		chatID, messageID, cancelledBy,
	).Scan(&targetUserID)
	return targetUserID, err
}

func (s *Storage) ReinstateWorkout(ctx context.Context, chatID int64, messageID int, reinstatedBy int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE workouts 
		SET is_cancelled = false, cancelled_by = NULL, cancelled_at = NULL 
		WHERE chat_id = $1 AND completion_message_id = $2 AND cancelled_by = $3`,
		chatID, messageID, reinstatedBy,
	)
	return err
}

func (s *Storage) GetWorkoutByMessage(ctx context.Context, chatID int64, messageID int) (*Workout, error) {
	var w Workout
	err := s.db.QueryRowContext(ctx, `
		SELECT id, workout_date, user_id, chat_id, completion_message_id, is_cancelled, cancelled_by, cancelled_at
		FROM workouts 
		WHERE chat_id = $1 AND completion_message_id = $2`,
		chatID, messageID,
	).Scan(&w.ID, &w.WorkoutDate, &w.UserID, &w.ChatID, &w.MessageID, &w.IsCancelled, &w.CancelledBy, &w.CancelledAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Storage) AddWorkouts(ctx context.Context, userID int64, chatID int64, amount int) (int, error) {
	// To "add" a workout, we create new workout records for the current date.
	count := 0
	for i := 0; i < amount; i++ {
		query := `INSERT INTO workouts (user_id, workout_date, chat_id, completion_message_id)
		VALUES ($1, NOW(), $2, 0)`
		_, err := s.db.ExecContext(ctx, query, userID, chatID)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *Storage) SubtractWorkouts(ctx context.Context, userID int64, chatID int64, amount int) (int, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE workouts 
		SET is_cancelled = true, cancelled_by = 0, cancelled_at = NOW()
		WHERE id IN (
			SELECT id FROM workouts 
			WHERE user_id = $1 AND chat_id = $2 AND is_cancelled = false 
			ORDER BY workout_date DESC 
			LIMIT $3
		)`,
		userID, chatID, amount,
	)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	return int(affected), err
}
