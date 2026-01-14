package auth

import (
	"context"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//認証

type Usecase struct {
	Users      repo.UserRepository
	Sessions   repo.SessionRepository
	SessionTTL time.Duration
}

type LoginResult struct {
	UserID    string
	Email     string
	SessionID string
	ExpiresAt time.Time
}

type MeResult struct {
	UserID string
	Email  string
}

// ログイン
func (u Usecase) Login(ctx context.Context, email string, password string) (LoginResult, error) {
	//ユーザー検索
	user, err := u.Users.FindByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, domain.ErrUnauthorized
	}

	//パスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, domain.ErrUnauthorized
	}
	//セッション作成でDBに保存
	expires := time.Now().Add(u.SessionTTL)
	sess, err := u.Sessions.Create(ctx, user.ID, expires)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		UserID:    user.ID,
		Email:     user.Email,
		SessionID: sess.ID,
		ExpiresAt: sess.ExpiresAt,
	}, nil
}

func (u Usecase) Me(ctx context.Context, sessionID string) (MeResult, error) {
	sess, err := u.Sessions.FindByID(ctx, sessionID)
	if err != nil {
		return MeResult{}, domain.ErrUnauthorized
	}
	//期限切れ
	if time.Now().After(sess.ExpiresAt) {
		return MeResult{}, domain.ErrUnauthorized
	}

	user, err := u.Users.FindByID(ctx, sess.UserID)
	if err != nil {
		return MeResult{}, domain.ErrUnauthorized
	}
	return MeResult{UserID: user.ID, Email: user.Email}, nil
}

// ログアウト
func (u Usecase) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return u.Sessions.Delete(ctx, sessionID)
}
