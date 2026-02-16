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
	query := `INSERT INTO challenges (is_active, challenge_duration)
	VALUES ($1, $2)
	
	`

	_, err := s.db.ExecContext(ctx, query, challenge.IsActive, challenge.Duration)

	return err
}

func (s *Storage) GetChallenge(ctx context.Context, challengeID int) (*Challenge, error) {
	var challenge Challenge
	query := `SELECT id, days_per_week, challenge_duration, is_active
			  FROM challenges
			  WHERE id = $1
			  `
	err := s.db.QueryRowContext(ctx, query).Scan(
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
			  SET challenge_duration = $1, is_active = $
			  WHERE id = $3`
	_, err := s.db.ExecContext(ctx, query,
		challenge.DaysPerWeek,
		challenge.Duration,
		challenge.IsActive,
		challenge.ChallengeID,
	)

	if err != nil {
		return err
	}
	return nil

}

func (s *Storage) SetWeekRules(ctx context.Context, challengeID int, days int) error {
	query := `
        UPDATE challenges 
        SET days_per_week = $1 
        WHERE id = $2 AND days_per_week = 0`

	result, err := s.db.ExecContext(ctx, query)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("the rule has already been set for this challenge")
	}

	return nil
}
