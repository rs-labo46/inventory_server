package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"inventory-backend/internal/application/auth"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"

	"golang.org/x/crypto/bcrypt"
)

type userRepoStub struct {
	findByEmail func(ctx context.Context, email string) (domain.User, error)
	findByID    func(ctx context.Context, id string) (domain.User, error)
	create      func(ctx context.Context, email string, passwordHash string) (domain.User, error)
}

func (u userRepoStub) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return u.findByEmail(ctx, email)
}
func (u userRepoStub) FindByID(ctx context.Context, userID string) (domain.User, error) {
	return u.findByID(ctx, userID)
}

func (u userRepoStub) Create(ctx context.Context, email string, passwordHash string) (domain.User, error) {

	if u.create == nil {
		return domain.User{}, errors.New("Create is not stubbed")
	}
	return u.create(ctx, email, passwordHash)
}

var _ repo.UserRepository = userRepoStub{}

type sessionRepoStub struct {
	create   func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error)
	findByID func(ctx context.Context, sessionID string) (domain.Session, error)
	delete   func(ctx context.Context, sessionID string) error
}

func (s sessionRepoStub) Create(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
	return s.create(ctx, userID, expiresAt)
}
func (s sessionRepoStub) FindByID(ctx context.Context, sessionID string) (domain.Session, error) {
	return s.findByID(ctx, sessionID)
}
func (s sessionRepoStub) Delete(ctx context.Context, sessionID string) error {
	return s.delete(ctx, sessionID)
}

var _ repo.SessionRepository = sessionRepoStub{}

func TestAuthUsecase_Login(t *testing.T) {
	t.Parallel()

	t.Run("FindByEmail失敗は ErrUnauthorized", func(t *testing.T) {
		t.Parallel()

		uc := auth.Usecase{
			Users: userRepoStub{
				findByEmail: func(ctx context.Context, email string) (domain.User, error) {
					return domain.User{}, errors.New("not found")
				},
				findByID: func(ctx context.Context, userID string) (domain.User, error) {
					return domain.User{}, errors.New("not used")
				},
			},
			Sessions: sessionRepoStub{
				create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
					return domain.Session{}, nil
				},
				findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
					return domain.Session{}, nil
				},
				delete: func(ctx context.Context, sessionID string) error { return nil },
			},
			SessionTTL: 8 * time.Hour,
		}

		_, err := uc.Login(context.Background(), "x@example.com", "pw")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrUnauthorized)
		}
	})

	t.Run("パスワード不一致は ErrUnauthorized", func(t *testing.T) {
		t.Parallel()

		hash, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("bcrypt error: %v", err)
		}

		uc := auth.Usecase{
			Users: userRepoStub{
				findByEmail: func(ctx context.Context, email string) (domain.User, error) {
					return domain.User{
						ID:           "u1",
						Email:        email,
						PasswordHash: string(hash),
					}, nil
				},
				findByID: func(ctx context.Context, userID string) (domain.User, error) {
					return domain.User{}, errors.New("not used")
				},
			},
			Sessions: sessionRepoStub{
				create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
					return domain.Session{}, nil
				},
				findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
					return domain.Session{}, nil
				},
				delete: func(ctx context.Context, sessionID string) error { return nil },
			},
			SessionTTL: 8 * time.Hour,
		}

		_, gotErr := uc.Login(context.Background(), "x@example.com", "wrong")
		if !errors.Is(gotErr, domain.ErrUnauthorized) {
			t.Fatalf("error mismatch: got=%v want=%v", gotErr, domain.ErrUnauthorized)
		}
	})

	t.Run("成功: session作成して結果を返す", func(t *testing.T) {
		t.Parallel()

		hash, err := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("bcrypt error: %v", err)
		}

		var createdUserID string
		var createdExpires time.Time

		uc := auth.Usecase{
			Users: userRepoStub{
				findByEmail: func(ctx context.Context, email string) (domain.User, error) {
					return domain.User{
						ID:           "u1",
						Email:        email,
						PasswordHash: string(hash),
					}, nil
				},
				findByID: func(ctx context.Context, userID string) (domain.User, error) {
					return domain.User{}, errors.New("not used")
				},
			},
			Sessions: sessionRepoStub{
				create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
					createdUserID = userID
					createdExpires = expiresAt
					return domain.Session{ID: "s1", UserID: userID, ExpiresAt: expiresAt}, nil
				},
				findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
					return domain.Session{}, errors.New("not used")
				},
				delete: func(ctx context.Context, sessionID string) error { return nil },
			},
			SessionTTL: 2 * time.Hour,
		}

		got, err := uc.Login(context.Background(), "x@example.com", "pw")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if createdUserID != "u1" {
			t.Fatalf("create userID mismatch: got=%s want=%s", createdUserID, "u1")
		}
		if createdExpires.IsZero() {
			t.Fatalf("expiresAt should be set")
		}

		if got.UserID != "u1" || got.SessionID != "s1" || got.Email != "x@example.com" {
			t.Fatalf("login result mismatch: %+v", got)
		}
	})
}

func TestAuthUsecase_Me(t *testing.T) {
	t.Parallel()

	t.Run("セッションが無ければ ErrUnauthorized", func(t *testing.T) {
		t.Parallel()

		uc := auth.Usecase{
			Users: userRepoStub{
				findByEmail: func(ctx context.Context, email string) (domain.User, error) { return domain.User{}, nil },
				findByID:    func(ctx context.Context, userID string) (domain.User, error) { return domain.User{}, nil },
			},
			Sessions: sessionRepoStub{
				findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
					return domain.Session{}, errors.New("not found")
				},
				create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
					return domain.Session{}, nil
				},
				delete: func(ctx context.Context, sessionID string) error { return nil },
			},
			SessionTTL: 1 * time.Hour,
		}

		_, err := uc.Me(context.Background(), "s1")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrUnauthorized)
		}
	})

	t.Run("期限切れなら ErrUnauthorized", func(t *testing.T) {
		t.Parallel()

		uc := auth.Usecase{
			Users: userRepoStub{
				findByEmail: func(ctx context.Context, email string) (domain.User, error) { return domain.User{}, nil },
				findByID:    func(ctx context.Context, userID string) (domain.User, error) { return domain.User{}, nil },
			},
			Sessions: sessionRepoStub{
				findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
					return domain.Session{ID: sessionID, UserID: "u1", ExpiresAt: time.Now().Add(-1 * time.Minute)}, nil
				},
				create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
					return domain.Session{}, nil
				},
				delete: func(ctx context.Context, sessionID string) error { return nil },
			},
			SessionTTL: 1 * time.Hour,
		}

		_, err := uc.Me(context.Background(), "s1")
		if !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrUnauthorized)
		}
	})

	t.Run("成功: session→user を引ける", func(t *testing.T) {
		t.Parallel()

		uc := auth.Usecase{
			Users: userRepoStub{
				findByEmail: func(ctx context.Context, email string) (domain.User, error) { return domain.User{}, nil },
				findByID: func(ctx context.Context, userID string) (domain.User, error) {
					return domain.User{ID: userID, Email: "x@example.com"}, nil
				},
			},
			Sessions: sessionRepoStub{
				findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
					return domain.Session{ID: sessionID, UserID: "u1", ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
				},
				create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
					return domain.Session{}, nil
				},
				delete: func(ctx context.Context, sessionID string) error { return nil },
			},
			SessionTTL: 1 * time.Hour,
		}

		got, err := uc.Me(context.Background(), "s1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.UserID != "u1" || got.Email != "x@example.com" {
			t.Fatalf("me result mismatch: %+v", got)
		}
	})
}

func TestAuthUsecase_Logout(t *testing.T) {
	t.Parallel()

	t.Run("空sessionIDなら何もしない", func(t *testing.T) {
		t.Parallel()

		called := false

		uc := auth.Usecase{
			Sessions: sessionRepoStub{
				delete: func(ctx context.Context, sessionID string) error {
					called = true
					return nil
				},
				create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
					return domain.Session{}, nil
				},
				findByID: func(ctx context.Context, sessionID string) (domain.Session, error) { return domain.Session{}, nil },
			},
		}

		if err := uc.Logout(context.Background(), ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if called {
			t.Fatalf("Delete should not be called when sessionID is empty")
		}
	})

	t.Run("sessionIDがあればDeleteが呼ばれる", func(t *testing.T) {
		t.Parallel()

		called := false

		uc := auth.Usecase{
			Sessions: sessionRepoStub{
				delete: func(ctx context.Context, sessionID string) error {
					called = true
					if sessionID != "s1" {
						t.Fatalf("sessionID mismatch: got=%s want=%s", sessionID, "s1")
					}
					return nil
				},
				create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
					return domain.Session{}, nil
				},
				findByID: func(ctx context.Context, sessionID string) (domain.Session, error) { return domain.Session{}, nil },
			},
		}

		if err := uc.Logout(context.Background(), "s1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("Delete was not called")
		}
	})
}
