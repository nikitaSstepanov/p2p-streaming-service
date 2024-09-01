package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/dto"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

const (
	ok = http.StatusOK
	badReq = http.StatusBadRequest
	unauth = http.StatusUnauthorized

	cookieName = "refreshToken"
	cookieAge  = 259200
	cookiePath = "/"
	cookieHost = "localhost"
)

var (
	badReqErr  = e.New("Incorrect data.", e.BadInput)
	unauthErr  = e.New("You are unauth.", e.Unauthorize)

	logoutMsg  = responses.NewMessage("Logout success.")
)

type AuthUseCase interface {
	Login(ctx context.Context, user *entity.User) (*entity.Tokens, *e.Error)
	Logout(ctx context.Context, userId uint64) *e.Error
	Refresh(ctx context.Context, token string) (*entity.Tokens, *e.Error)
}

type Auth struct {
	usecase AuthUseCase
}

func New(usecase AuthUseCase) *Auth {
	return &Auth{
		usecase,
	}
}

func (a *Auth) Login(ctx *gin.Context) {
	var body dto.LoginDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	tokens, err := a.usecase.Login(ctx, body.ToEntity())
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return
	}

	ctx.SetCookie(cookieName, tokens.Refresh, cookieAge, cookiePath, cookieHost, false, true)

	ctx.JSON(ok, dto.TokenToDto(tokens))
}

func (a *Auth) Logout(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	err := a.usecase.Logout(ctx, userId)
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return 
	}

	ctx.SetCookie(cookieName, "", -1, cookiePath, cookieHost, false, true)

	ctx.JSON(ok, logoutMsg)
}

func (a *Auth) Refresh(ctx *gin.Context) {
	refresh, cookieErr := ctx.Cookie(cookieName)
	if cookieErr != nil {
		ctx.AbortWithStatusJSON(unauth, unauthErr)
		return 
	}

	tokens, err := a.usecase.Refresh(ctx, refresh)
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return 
	}

	ctx.SetCookie(cookieName, tokens.Refresh, cookieAge, cookiePath, cookieHost, false, true)

	ctx.JSON(ok, dto.TokenToDto(tokens))
}