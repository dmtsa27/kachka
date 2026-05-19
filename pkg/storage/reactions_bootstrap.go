package storage

import "context"

func (s *Storage) InitChallengeBootstrap(ctx context.Context, chatID int64, welcomeMessageID int, isBotAdmin bool, expectedReactions int) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO challenge_bootstrap (chat_id, welcome_message_id, expected_reactions, roster_frozen_at, is_started, started_at, is_bot_admin, days_per_week, duration_days, price)
		VALUES ($1, $2, $3, NOW(), false, NULL, $4, 3, 180, 500)
		ON CONFLICT (chat_id)
		DO UPDATE SET
			welcome_message_id = EXCLUDED.welcome_message_id,
			expected_reactions = EXCLUDED.expected_reactions,
			roster_frozen_at = NOW(),
			is_started = false,
			started_at = NULL,
			is_bot_admin = EXCLUDED.is_bot_admin
	`, chatID, welcomeMessageID, expectedReactions, isBotAdmin)
	return err
}

func (s *Storage) GetChallengeBootstrap(ctx context.Context, chatID int64) (*ChallengeBootstrap, error) {
	var state ChallengeBootstrap
	err := s.db.QueryRowContext(ctx, `
		SELECT chat_id, welcome_message_id, expected_reactions, roster_frozen_at, is_started, started_at, is_bot_admin, days_per_week, duration_days, price
		FROM challenge_bootstrap
		WHERE chat_id = $1
	`, chatID).Scan(
		&state.ChatID,
		&state.WelcomeMessageID,
		&state.ExpectedReactions,
		&state.RosterFrozenAt,
		&state.IsStarted,
		&state.StartedAt,
		&state.IsBotAdmin,
		&state.DaysPerWeek,
		&state.DurationDays,
		&state.Price,
	)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *Storage) SetBotAdminStatus(ctx context.Context, chatID int64, isBotAdmin bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE challenge_bootstrap
		SET is_bot_admin = $2
		WHERE chat_id = $1
	`, chatID, isBotAdmin)
	return err
}

func (s *Storage) MarkChallengeStarted(ctx context.Context, chatID int64) (bool, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE challenge_bootstrap
		SET is_started = true, started_at = NOW()
		WHERE chat_id = $1 AND is_started = false
	`, chatID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *Storage) UpdateChallengeBootstrapConfig(ctx context.Context, chatID int64, daysPerWeek int, durationDays int, price int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE challenge_bootstrap
		SET days_per_week = $2, duration_days = $3, price = $4
		WHERE chat_id = $1
	`, chatID, daysPerWeek, durationDays, price)
	return err
}

func (s *Storage) UpsertChatMember(ctx context.Context, chatID int64, userID int64, isBot bool, isActive bool) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_members (chat_id, user_id, is_bot, is_active, first_seen_at, last_seen_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (chat_id, user_id)
		DO UPDATE SET
			is_bot = EXCLUDED.is_bot,
			is_active = EXCLUDED.is_active,
			last_seen_at = NOW()
	`, chatID, userID, isBot, isActive)
	return err
}

func (s *Storage) SetUserMessageReactions(ctx context.Context, chatID int64, messageID int, userID int64, emojis []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM message_reactions
		WHERE chat_id = $1 AND message_id = $2 AND user_id = $3
	`, chatID, messageID, userID); err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(emojis))
	for _, emoji := range emojis {
		if emoji == "" {
			continue
		}
		if _, ok := seen[emoji]; ok {
			continue
		}
		seen[emoji] = struct{}{}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO message_reactions (chat_id, message_id, user_id, emoji, is_active, updated_at)
			VALUES ($1, $2, $3, $4, true, NOW())
		`, chatID, messageID, userID, emoji); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) CountFrozenMembers(ctx context.Context, chatID int64) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM chat_members cm
		JOIN challenge_bootstrap cb ON cb.chat_id = cm.chat_id
		WHERE cm.chat_id = $1
		  AND cm.is_bot = false
		  AND cm.is_active = true
		  AND cm.first_seen_at <= cb.roster_frozen_at
	`, chatID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Storage) CountWelcomeHeartReactions(ctx context.Context, chatID int64) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT mr.user_id)
		FROM challenge_bootstrap cb
		JOIN message_reactions mr
		  ON mr.chat_id = cb.chat_id
		 AND mr.message_id = cb.welcome_message_id
		 AND mr.is_active = true
		 AND mr.emoji IN ('❤', '❤️')
		WHERE cb.chat_id = $1
	`, chatID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Storage) CancelCountedByMessage(ctx context.Context, chatID int64, messageID int) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	sessionResult, err := tx.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE chat_id = $1 AND message_id = $2 AND session_date = CURRENT_DATE
	`, chatID, messageID)
	if err != nil {
		return false, err
	}

	sessionRows, err := sessionResult.RowsAffected()
	if err != nil {
		return false, err
	}
	if sessionRows > 0 {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return true, nil
	}

	workoutResult, err := tx.ExecContext(ctx, `
		DELETE FROM workouts
		WHERE chat_id = $1 AND completion_message_id = $2 AND CAST(workout_date AS DATE) = CURRENT_DATE
	`, chatID, messageID)
	if err != nil {
		return false, err
	}

	workoutRows, err := workoutResult.RowsAffected()
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return workoutRows > 0, nil
}
