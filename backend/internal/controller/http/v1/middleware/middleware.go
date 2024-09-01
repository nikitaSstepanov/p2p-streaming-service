package middleware

import (
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecase/pkg/auth"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

type JwtUseCase interface {
	ValidateToken(jwtString string) (*auth.Claims, *e.Error)
}

type Middleware struct {
	jwt JwtUseCase
}

func New(uc JwtUseCase) *Middleware {
	return &Middleware{
		jwt: uc,
	}
}