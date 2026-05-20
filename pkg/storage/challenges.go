package storage

import (
	"context"
	"fmt"
)

func (s *Storage) CreateChallenge(ctx context.Context, challenge Challenge) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE challenges SET is_active = false WHERE chat_id = $1 AND is_active = true`, challenge.ChatID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO challenges (chat_id, days_per_week, challenge_duration, is_active, price, started_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())`,
		challenge.ChatID,
		challenge.DaysPerWeek,
		challenge.Duration,
		challenge.IsActive,
		challenge.Price,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) HasActiveChallengeInChat(ctx context.Context, chatID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM challenges WHERE chat_id = $1 AND is_active = true)`,
		chatID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Storage) GetChallenge(ctx context.Context, challengeID int) (*Challenge, error) {
	var challenge Challenge
	query := `SELECT id, days_per_week, challenge_duration, is_active, price
              FROM challenges
              WHERE id = $1`

	err := s.db.QueryRowContext(ctx, query, challengeID).Scan(
		&challenge.ChallengeID,
		&challenge.DaysPerWeek,
		&challenge.Duration,
		&challenge.IsActive,
		&challenge.Price,
	)

	if err != nil {
		return nil, err
	}

	return &challenge, nil
}

func (s *Storage) UpdateChallenge(ctx context.Context, challenge Challenge) error {
	query := `UPDATE challenges
              SET challenge_duration = $1, is_active = $2, price = $3
              WHERE id = $4`

	_, err := s.db.ExecContext(ctx, query,
		challenge.Duration,
		challenge.IsActive,
		challenge.Price,
		challenge.ChallengeID,
	)

	return err
}

func (s *Storage) GetActiveChallenge(ctx context.Context) (*Challenge, error) {
	var challenge Challenge
	query := `SELECT id, days_per_week, challenge_duration, is_active, price, started_at, chat_id, last_weekly_check_at, last_daily_stats_at
              FROM challenges WHERE is_active = true LIMIT 1`

	err := s.db.QueryRowContext(ctx, query).Scan(
		&challenge.ChallengeID,
		&challenge.DaysPerWeek,
		&challenge.Duration,
		&challenge.IsActive,
		&challenge.Price,
		&challenge.StartedAt,
		&challenge.ChatID,
		&challenge.LastWeeklyCheckAt,
		&challenge.LastDailyStatsAt,
	)
	if err != nil {
		return nil, err
	}

	return &challenge, nil
}

func (s *Storage) GetActiveChallengeByChat(ctx context.Context, chatID int64) (*Challenge, error) {
	var challenge Challenge
	query := `SELECT id, days_per_week, challenge_duration, is_active, price, started_at, chat_id, last_weekly_check_at, last_daily_stats_at
              FROM challenges WHERE chat_id = $1 AND is_active = true LIMIT 1`

	err := s.db.QueryRowContext(ctx, query, chatID).Scan(
		&challenge.ChallengeID,
		&challenge.DaysPerWeek,
		&challenge.Duration,
		&challenge.IsActive,
		&challenge.Price,
		&challenge.StartedAt,
		&challenge.ChatID,
		&challenge.LastWeeklyCheckAt,
		&challenge.LastDailyStatsAt,
	)
	if err != nil {
		return nil, err
	}

	return &challenge, nil
}

func (s *Storage) GetAllActiveChallenges(ctx context.Context) ([]Challenge, error) {
	query := `SELECT id, days_per_week, challenge_duration, is_active, price, started_at, chat_id, last_weekly_check_at, last_daily_stats_at
              FROM challenges WHERE is_active = true`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []Challenge
	for rows.Next() {
		var challenge Challenge
		err := rows.Scan(
			&challenge.ChallengeID,
			&challenge.DaysPerWeek,
			&challenge.Duration,
			&challenge.IsActive,
			&challenge.Price,
			&challenge.StartedAt,
			&challenge.ChatID,
			&challenge.LastWeeklyCheckAt,
			&challenge.LastDailyStatsAt,
		)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}

	return challenges, rows.Err()
}

func (s *Storage) MarkWeeklyCheckDone(ctx context.Context, challengeID int) error {
	query := `UPDATE challenges SET last_weekly_check_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, challengeID)
	return err
}

func (s *Storage) MarkDailyStatsDone(ctx context.Context, challengeID int) error {
	query := `UPDATE challenges SET last_daily_stats_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, challengeID)
	return err
}

// DeactivateChallengeForChat деактивує активний челендж та очищує дані для конкретного чату.
// Викликається коли бота кікають або чат видаляється.
func (s *Storage) DeleteChallenge(ctx context.Context, challengeID int) error {
	query := `DELETE FROM challenges WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, challengeID)
	return err
}

func (s *Storage) DeactivateChallengeForChat(ctx context.Context, chatID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Deactivate active challenges
	if _, err := tx.ExecContext(ctx, `UPDATE challenges SET is_active = false WHERE chat_id = $1`, chatID); err != nil {
		return err
	}

	// Delete bootstrap state
	if _, err := tx.ExecContext(ctx, `DELETE FROM challenge_bootstrap WHERE chat_id = $1`, chatID); err != nil {
		return err
	}

	// Delete message reactions
	if _, err := tx.ExecContext(ctx, `DELETE FROM message_reactions WHERE chat_id = $1`, chatID); err != nil {
		return err
	}

	// Delete chat members
	if _, err := tx.ExecContext(ctx, `DELETE FROM chat_members WHERE chat_id = $1`, chatID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) SetWeekRules(ctx context.Context, challengeID int, days int) error {
	query := `
        UPDATE challenges 
        SET days_per_week = $1 
        WHERE id = $2 AND days_per_week = 0`

	result, err := s.db.ExecContext(ctx, query, days, challengeID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("the rule has already been set for this challenge")
	}

	return nil
}
