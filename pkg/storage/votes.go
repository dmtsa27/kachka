package storage

import (
	"context"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

func (s *Storage) CreateVote(ctx context.Context, vote domain.Vote) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO votes (chat_id, target_user_id, initiator_id, poll_id, amount, expires_at, type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		vote.ChatID, vote.TargetUserID, vote.InitiatorID, vote.PollID, vote.Amount, vote.ExpiresAt, vote.Type,
	)
	return err
}

func (s *Storage) GetVoteByPollID(ctx context.Context, pollID string) (*domain.Vote, error) {
	var v domain.Vote
	err := s.db.QueryRowContext(ctx, `
		SELECT id, chat_id, target_user_id, initiator_id, poll_id, amount, created_at, expires_at, is_completed, is_success, type
		FROM votes WHERE poll_id = $1`,
		pollID,
	).Scan(&v.ID, &v.ChatID, &v.TargetUserID, &v.InitiatorID, &v.PollID, &v.Amount, &v.CreatedAt, &v.ExpiresAt, &v.IsCompleted, &v.IsSuccess, &v.Type)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *Storage) CompleteVote(ctx context.Context, pollID string, success bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE votes SET is_completed = true, is_success = $2 WHERE poll_id = $1`,
		pollID, success,
	)
	return err
}
