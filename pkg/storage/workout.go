package storage

import (
	"context"
	"time"

	"github.com/dmtsa27/kachka.git/pkg/domain"
	"github.com/dmtsa27/kachka.git/pkg/service"
)

func (s *Storage) CreateWorkout(ctx context.Context, workout Workout) error {
	query := `INSERT INTO workouts (user_id, workout_date, chat_id, completion_message_id)
	VALUES ($1, $2, $3, $4)`
	_, err := s.db.ExecContext(ctx, query, workout.UserID, workout.WorkoutDate, workout.ChatID, workout.MessageID)
	return err
}

func (s *Storage) WeeklyWorkouts(ctx context.Context, userID int64, chatID int64, weekStart time.Time) (int, error) {
	query := `
        SELECT COUNT(DISTINCT CAST(workout_date AS DATE)) 
        FROM workouts 
        WHERE user_id = $1 
          AND chat_id = $2
          AND workout_date >= $3
          AND workout_date < $3 + INTERVAL '7 days'
          AND is_cancelled = false`

	var count int
	err := s.db.QueryRowContext(ctx, query, userID, chatID, weekStart).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Storage) GetWorkoutCounts(ctx context.Context, chatID int64, weekStart time.Time) ([]service.UserWorkouts, error) {
	query := `
        SELECT u.telegram_id, u.username, 
               (SELECT COUNT(DISTINCT CAST(w.workout_date AS DATE))
                FROM workouts w
                WHERE w.user_id = u.telegram_id
                  AND w.chat_id = $1
                  AND w.workout_date >= $2 
                  AND w.workout_date < $2 + INTERVAL '7 days'
                  AND w.is_cancelled = false) as workout_count
        FROM chat_members cm
        JOIN users u ON u.telegram_id = cm.user_id
        WHERE cm.chat_id = $1 AND cm.is_active = true AND cm.is_bot = false AND u.is_active = true`

	rows, err := s.db.QueryContext(ctx, query, chatID, weekStart)
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
	now := time.Now().UTC()
	query := `
		SELECT EXISTS(
			SELECT 1 FROM workouts 
			WHERE user_id = $1 AND chat_id = $2 
			AND CAST(workout_date AS DATE) = CAST($3 AS DATE) 
			AND is_cancelled = false
		)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userID, chatID, now).Scan(&exists)
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
	added := 0
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)

	for i := 0; added < amount && i < 14; i++ {
		date := today.AddDate(0, 0, -i)
		query := `
			INSERT INTO workouts (user_id, workout_date, chat_id, completion_message_id)
			SELECT $1, $2, $3, 0
			WHERE NOT EXISTS (
				SELECT 1 FROM workouts 
				WHERE user_id = $1 AND chat_id = $3 
				AND CAST(workout_date AS DATE) = CAST($2 AS DATE)
				AND is_cancelled = false
			)`
		res, err := s.db.ExecContext(ctx, query, userID, date, chatID)
		if err != nil {
			return added, err
		}
		rows, _ := res.RowsAffected()
		if rows > 0 {
			added++
		}
	}
	return added, nil
}

func (s *Storage) SubtractWorkouts(ctx context.Context, userID int64, chatID int64, amount int) (int, error) {
	// We want to subtract active workouts, starting from the most recent ones.
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
	return int(affected), nil
}

func (s *Storage) GetChatStats(ctx context.Context, chatID int64, weekStart time.Time) ([]domain.UserStats, error) {
	query := `
        SELECT 
            u.telegram_id, 
            COALESCE(u.username, ''), 
            (SELECT COUNT(DISTINCT CAST(w.workout_date AS DATE)) 
             FROM workouts w 
             WHERE w.user_id = u.telegram_id AND w.chat_id = $1 
               AND w.workout_date >= $2 AND w.workout_date < $2 + INTERVAL '7 days'
               AND w.is_cancelled = false) as weekly_count,
            (SELECT COUNT(DISTINCT CAST(w.workout_date AS DATE)) 
             FROM workouts w 
             WHERE w.user_id = u.telegram_id AND w.chat_id = $1 
               AND w.is_cancelled = false) as total_count,
            u.is_active,
            EXISTS(
                SELECT 1 FROM workouts w_today 
                WHERE w_today.user_id = u.telegram_id 
                  AND w_today.chat_id = $1 
                  AND CAST(w_today.workout_date AS DATE) = CAST($3 AS DATE) 
                  AND w_today.is_cancelled = false
            ) as has_workout_today
        FROM chat_members cm
        JOIN users u ON u.telegram_id = cm.user_id
        WHERE cm.chat_id = $1 AND cm.is_bot = false AND cm.is_active = true
        ORDER BY weekly_count DESC, total_count DESC`

	rows, err := s.db.QueryContext(ctx, query, chatID, weekStart, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.UserStats
	for rows.Next() {
		var us domain.UserStats
		if err := rows.Scan(&us.TelegramID, &us.Username, &us.WeeklyCount, &us.TotalCount, &us.IsActive, &us.HasWorkoutToday); err != nil {
			return nil, err
		}
		stats = append(stats, us)
	}

	return stats, rows.Err()
}

func (s *Storage) GetActiveChallengeVotersCount(ctx context.Context, chatID int64) (int, error) {
	var count int
	query := `
        SELECT COUNT(*)
        FROM chat_members cm
        JOIN users u ON u.telegram_id = cm.user_id
        WHERE cm.chat_id = $1
          AND cm.is_active = true
          AND cm.is_bot = false
          AND u.is_active = true`
	err := s.db.QueryRowContext(ctx, query, chatID).Scan(&count)
	return count, err
}
