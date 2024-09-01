package auth

import (
	"context"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessExpires  = 1 * time.Hour
	refreshExpires = 72 * time.Hour
)

var (
	loginErr   = e.New("Incorrect username or password", e.Unauthorize)
	refreshErr = e.New("Invalid refresh", e.Unauthorize)
)

type UserStorage interface {
	GetUser(ctx context.Context, id uint64) (*entity.User, *e.Error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, *e.Error)
	GetUsersByRole(ctx context.Context, role string) ([]*entity.User, *e.Error)
	Update(ctx context.Context, user *entity.User) *e.Error
}

type TokenStorage interface {
	GetUserByRefresh(ctx context.Context, refresh string) (*entity.User, *e.Error)
	SetRefresh(ctx context.Context, userId uint64, token string) *e.Error
	DeleteRefresh(ctx context.Context, userId uint64) *e.Error
}

type Auth struct {
	jwt    *JwtUseCase
	users  UserStorage
	tokens TokenStorage
}

func New(jwt *JwtUseCase, users UserStorage, tokens TokenStorage) *Auth {
	return &Auth{
		jwt,
		users,
		tokens,
	}
}

func (a *Auth) Login(ctx context.Context, user *entity.User) (*entity.Tokens, *e.Error) {
	candidate, err := a.users.GetUserByUsername(ctx, user.Username)
	if err != nil {
		return nil, err
	}

	if candidate.Id == 0 {
		return nil, loginErr
	}

	checkPasswordErr := a.checkPassword(candidate.Password, user.Password)
	if checkPasswordErr != nil {
		return nil, loginErr
	}

	access, err := a.jwt.GenerateToken(candidate.Id, candidate.Role, accessExpires)
	if err != nil {
		return nil, err
	}

	refresh, err := a.jwt.GenerateToken(candidate.Id, candidate.Role, refreshExpires)
	if err != nil {
		return nil, err
	}

	err = a.tokens.SetRefresh(ctx, user.Id, refresh)
	if err != nil {
		return nil, err
	}

	result := &entity.Tokens{
		Access: access,
		Refresh: refresh,
	}

	return result, nil
}

func (a *Auth) Logout(ctx context.Context, userId uint64) *e.Error {
	return a.tokens.DeleteRefresh(ctx, userId)
}

func (a *Auth) Refresh(ctx context.Context, token string) (*entity.Tokens, *e.Error) {
	claims, err := a.jwt.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	candidate, err := a.tokens.GetUserByRefresh(ctx, token)
	if err != nil {
		return nil, err
	}

	if candidate.Id == 0 || candidate.Id != claims.Id {
		return nil, refreshErr
	}

	access, err := a.jwt.GenerateToken(candidate.Id, candidate.Role, accessExpires)
	if err != nil {
		return nil, err
	}

	refresh, err := a.jwt.GenerateToken(candidate.Id, candidate.Role, refreshExpires)
	if err != nil {
		return nil, err
	}

	err = a.tokens.SetRefresh(ctx, candidate.Id, refresh)
	if err != nil {
		return nil, err
	}

	result := &entity.Tokens{
		Access: access,
		Refresh: refresh,
	}

	return result, nil
}

func (a *Auth) checkPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}