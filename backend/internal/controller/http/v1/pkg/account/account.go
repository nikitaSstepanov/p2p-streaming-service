package account

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/dto"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
	"github.com/go-playground/validator/v10"
	"github.com/gin-gonic/gin"
)

const (
	ok         = http.StatusOK
	created    = http.StatusCreated
	badReq     = http.StatusBadRequest
	
	cookieName = "refreshToken"
	cookieAge  = 259200
	cookiePath = "/"
	cookieHost = "localhost"
)

var (
	badReqErr = e.New("Incorrect data.", e.BadInput)
)

type AccountUseCase interface {
	GetAccount(ctx context.Context, id uint64) (*entity.User, *e.Error)
	Create(ctx context.Context, user *entity.User) (*entity.Tokens, *e.Error)
} 

type Account struct {
	usecase AccountUseCase
}

func New(uc AccountUseCase) *Account {
	return &Account{
		usecase: uc,
	}
}

func (a *Account) GetAccount(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	account, err := a.usecase.GetAccount(ctx, userId)
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return
	}

	ctx.JSON(ok, dto.AccountToDto(account))
}

func (a *Account) Create(ctx *gin.Context) {
	var body dto.CreateUserDto
	
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	validate := validator.New()

	validErr := validate.Struct(body)
	if validErr != nil {
		ctx.AbortWithStatusJSON(badReq, getValidationError(validErr))
		return
	}

	tokens, err := a.usecase.Create(ctx, body.ToEntity())
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return 
	}

	ctx.SetCookie(cookieName, tokens.Refresh, cookieAge, cookiePath, cookieHost, false, true)

	ctx.JSON(created, dto.TokenToDto(tokens))
}

func getValidationError(err error) *e.Error {
	errors := err.(validator.ValidationErrors)
	msg := fmt.Sprintf("Incorrect data: %s", errors)
	return e.New(msg, e.BadInput)
}