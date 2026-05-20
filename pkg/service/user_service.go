package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dmtsa27/kachka.git/pkg/domain"
)

type userService struct {
	users UserRepository
}

func (u *userService) RegisterUser(ctx context.Context, telegramID int64, username string) error {
	return u.users.CreateUser(ctx, domain.User{
		TelegramID: telegramID,
		Username:   username,
		IsActive:   true,
	})
}

func (u *userService) ReadUser(ctx context.Context, telegramID int64) (*domain.User, error) {
	return u.users.ReadUser(ctx, telegramID)
}

func (u *userService) UpdateUser(ctx context.Context, user domain.User) error {
	return u.users.UpdateUser(ctx, user)
}

func (u *userService) DeleteUser(ctx context.Context, telegramID int64) error {
	return u.users.DeleteUser(ctx, telegramID)
}

func (u *userService) IsActiveUser(ctx context.Context, userID int64) (bool, error) {
	user, err := u.users.ReadUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("read user: %w", err)
	}
	return user.IsActive, nil
}

func (u *userService) GetUserIDByUsername(ctx context.Context, username string) (int64, error) {
	return u.users.GetUserIDByUsername(ctx, username)
}

func (u *userService) GetAllActiveUsers(ctx context.Context) ([]domain.User, error) {
	return u.users.GetAllActiveUsers(ctx)
}
