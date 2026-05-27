package storage

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	user := User{TelegramID: 123, Username: "testuser", IsActive: true}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.TelegramID, user.Username, user.IsActive).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.CreateUser(context.Background(), user); err != nil {
		t.Errorf("error was not expected while creating user: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestReadUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	now := time.Now()
	rows := sqlmock.NewRows([]string{"telegram_id", "username", "days_trained", "is_active", "failed_at"}).
		AddRow(123, "testuser", 5, true, now)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE telegram_id = ?").
		WithArgs(123).
		WillReturnRows(rows)

	user, err := s.ReadUser(context.Background(), 123)
	if err != nil {
		t.Errorf("error was not expected while reading user: %s", err)
	}

	if user.TelegramID != 123 || user.Username != "testuser" || user.DaysTrained != 5 || !user.IsActive {
		t.Errorf("unexpected user data: %+v", user)
	}
}

func TestDeactivateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := &Storage{db: db}

	mock.ExpectExec("UPDATE users SET is_active = false").
		WithArgs(123).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.DeactivateUser(context.Background(), 123); err != nil {
		t.Errorf("error was not expected while deactivating user: %s", err)
	}
}

func TestBatchDeactivateUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := &Storage{db: db}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET is_active = false").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE users SET is_active = false").WithArgs(2).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := s.BatchDeactivateUsers(context.Background(), []int64{1, 2}); err != nil {
		t.Errorf("error was not expected while batch deactivating users: %s", err)
	}
}

func TestGetAllActiveUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	rows := sqlmock.NewRows([]string{"telegram_id", "username", "days_trained", "is_active", "failed_at"}).
		AddRow(1, "user1", 1, true, nil).
		AddRow(2, "user2", 2, true, nil)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE is_active = true").WillReturnRows(rows)

	users, err := s.GetAllActiveUsers(context.Background())
	if err != nil {
		t.Errorf("error was not expected: %s", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestGetUserIDByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	mock.ExpectQuery("SELECT telegram_id FROM users WHERE username = ?").WithArgs("test").WillReturnRows(sqlmock.NewRows([]string{"telegram_id"}).AddRow(123))

	id, err := s.GetUserIDByUsername(context.Background(), "@test")
	if err != nil || id != 123 {
		t.Errorf("unexpected result: id=%d, err=%v", id, err)
	}
}

func TestDeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := &Storage{db: db}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM workouts").WithArgs(123).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM sessions").WithArgs(123).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM chat_members").WithArgs(123).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM message_reactions").WithArgs(123).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM users").WithArgs(123).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := s.DeleteUser(context.Background(), 123); err != nil {
		t.Errorf("error was not expected: %s", err)
	}
}

func TestCreateWorkout(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	w := Workout{UserID: 1, WorkoutDate: time.Now(), ChatID: 1, MessageID: 100}

	mock.ExpectExec("INSERT INTO workouts").
		WithArgs(w.UserID, w.WorkoutDate, w.ChatID, w.MessageID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.CreateWorkout(context.Background(), w); err != nil {
		t.Errorf("error: %s", err)
	}
}

func TestWeeklyWorkouts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	now := time.Now()

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(1, 1, now).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := s.WeeklyWorkouts(context.Background(), 1, 1, now)
	if err != nil || count != 3 {
		t.Errorf("unexpected result: %d, %v", count, err)
	}
}

func TestHasWorkoutToday(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}

	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := s.HasWorkoutToday(context.Background(), 1, 1)
	if err != nil || !exists {
		t.Errorf("unexpected result: %v, %v", exists, err)
	}
}

func TestCancelWorkout(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}

	mock.ExpectQuery("UPDATE workouts").
		WithArgs(1, 100, 2).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))

	targetUserID, err := s.CancelWorkout(context.Background(), 1, 100, 2)
	if err != nil || targetUserID != 1 {
		t.Errorf("unexpected result: %d, %v", targetUserID, err)
	}
}

func TestHasTrainedToday(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	trained, err := s.HasTrainedToday(context.Background(), 1, 1)
	if err != nil || !trained {
		t.Errorf("unexpected result: %v, %v", trained, err)
	}
}

func TestStartSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}

	mock.ExpectExec("INSERT INTO sessions").
		WithArgs(1, 1, 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.StartSession(context.Background(), 1, 1, 100); err != nil {
		t.Errorf("error: %v", err)
	}
}

func TestGetSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	now := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM sessions").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "chat_id", "started_at", "last_video_at"}).
			AddRow(1, 1, 1, now, now))

	sess, err := s.GetSession(context.Background(), 1, 1)
	if err != nil || sess.ID != 1 {
		t.Errorf("unexpected result: %v, %v", sess, err)
	}
}

func TestCreateChallenge(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	ch := Challenge{ChatID: 1, DaysPerWeek: 3, Duration: 30, IsActive: true, Price: 100}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE challenges").WithArgs(ch.ChatID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO challenges").
		WithArgs(ch.ChatID, ch.DaysPerWeek, ch.Duration, ch.IsActive, ch.Price).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := s.CreateChallenge(context.Background(), ch); err != nil {
		t.Errorf("error: %v", err)
	}
}

func TestHasActiveChallengeInChat(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}

	mock.ExpectQuery("SELECT EXISTS").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := s.HasActiveChallengeInChat(context.Background(), 1)
	if err != nil || !exists {
		t.Errorf("unexpected result: %v, %v", exists, err)
	}
}

func TestGetActiveChallenge(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer db.Close()

	s := &Storage{db: db}
	now := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM challenges WHERE is_active = true").
		WillReturnRows(sqlmock.NewRows([]string{"id", "days_per_week", "challenge_duration", "is_active", "price", "started_at", "chat_id", "last_weekly_check_at", "last_daily_stats_at"}).
			AddRow(1, 3, 30, true, 100, now, 1, now, now))

	ch, err := s.GetActiveChallenge(context.Background())
	if err != nil || ch.ChallengeID != 1 {
		t.Errorf("unexpected result: %v, %v", ch, err)
	}
}
