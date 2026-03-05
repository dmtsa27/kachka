package storage

import (
	"context"
	"fmt"
	"time"
)

type Challenge struct {
	ChallengeID int
	IsActive    bool
	DaysPerWeek int
	Duration    int
	StartedAt   time.Time
	ChatID      int64
}

func (s *Storage) CreateChallenge(ctx context.Context, challenge Challenge) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE challenges SET is_active = false WHERE is_active = true`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO challenges (chat_id, days_per_week, challenge_duration, is_active, started_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		challenge.ChatID,
		challenge.DaysPerWeek,
		challenge.Duration,
		challenge.IsActive,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetChallenge(ctx context.Context, challengeID int) (*Challenge, error) {
	var challenge Challenge
	query := `SELECT id, days_per_week, challenge_duration, is_active
              FROM challenges
              WHERE id = $1`

	err := s.db.QueryRowContext(ctx, query, challengeID).Scan(
		&challenge.ChallengeID,
		&challenge.DaysPerWeek,
		&challenge.Duration,
		&challenge.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &challenge, nil
}

func (s *Storage) UpdateChallenge(ctx context.Context, challenge Challenge) error {
	query := `UPDATE challenges
              SET challenge_duration = $1, is_active = $2
              WHERE id = $3`

	_, err := s.db.ExecContext(ctx, query,
		challenge.Duration,
		challenge.IsActive,
		challenge.ChallengeID,
	)

	return err
}

func (s *Storage) GetActiveChallenge(ctx context.Context) (*Challenge, error) {
	var challenge Challenge
	query := `SELECT id, days_per_week, challenge_duration, is_active, started_at, chat_id
              FROM challenges WHERE is_active = true LIMIT 1`

	err := s.db.QueryRowContext(ctx, query).Scan(
		&challenge.ChallengeID,
		&challenge.DaysPerWeek,
		&challenge.Duration,
		&challenge.IsActive,
		&challenge.StartedAt,
		&challenge.ChatID,
	)
	if err != nil {
		return nil, err
	}

	return &challenge, nil
}

// DeactivateChallengeForChat деактивує активний челендж для конкретного чату.
// Викликається коли бота кікають або чат видаляється.
func (s *Storage) DeactivateChallengeForChat(ctx context.Context, chatID int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE challenges SET is_active = false WHERE chat_id = $1 AND is_active = true`,
		chatID,
	)
	return err
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
