package account

import (
	"context"
	"time"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/auth"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessExpires  = 1 * time.Hour
	refreshExpires = 72 * time.Hour
	cost           = 10
)

var (
	internalErr = e.New("Something going wrong...", e.Internal)
)

type UserStorage interface {
	GetUser(ctx context.Context, id uint64) (*entity.User, *e.Error)
	Create(ctx context.Context, user *entity.User) *e.Error
}

type JwtUseCase interface {
	ValidateToken(jwtString string) (*auth.Claims, *e.Error)
	GenerateToken(id uint64, role string, expires time.Duration) (string, *e.Error)
}

type Account struct {
	storage UserStorage
	jwt     JwtUseCase
}

func New(storage UserStorage, jwt JwtUseCase) *Account {
	return &Account{
		storage,
		jwt,
	}
}

func (a *Account) GetAccount(ctx context.Context, id uint64) (*entity.User, *e.Error) {
	return a.storage.GetUser(ctx, id)
}

func (a *Account) Create(ctx context.Context, user *entity.User) (*entity.Tokens, *e.Error) {
	hash, hashPasswordErr := a.hashPassword(user.Password)
	if hashPasswordErr != nil {
		return nil, internalErr
	}

	user.Password = hash
	
	err := a.storage.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	var tokens entity.Tokens

	access, err := a.jwt.GenerateToken(user.Id, "USER", accessExpires)
	if err != nil {
		return nil, err
	}

	tokens.Access = access

	refresh, err := a.jwt.GenerateToken(user.Id, "USER", refreshExpires)
	if err != nil {
		return nil, err
	}

	tokens.Refresh = refresh

	return &tokens, nil
}

func (a *Account) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}