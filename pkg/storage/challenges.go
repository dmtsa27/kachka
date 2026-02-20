package storage

import (
	"context"
	"fmt"
)

type Challenge struct {
	ChallengeID int
	IsActive    bool
	DaysPerWeek int
	Duration    int
}

func (s *Storage) CreateChallenge(ctx context.Context, challenge Challenge) error {
	query := `INSERT INTO challenges (days_per_week, challenge_duration, is_active)
	VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, query,
		challenge.DaysPerWeek,
		challenge.Duration,
		challenge.IsActive,
	)

	return err
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
