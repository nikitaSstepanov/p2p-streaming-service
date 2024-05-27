package handlers

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/usecases/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/dto/users"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/types/statuses"
	"github.com/gin-gonic/gin"
)

type Account struct {
	services *services.Services
}

func NewAccount(services *services.Services) *Account {
	return &Account{
		services: services,
	}
}

func (a *Account) GetAccount(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	result, status := a.services.Account.GetAccount(ctx, header)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		a.handleError(ctx, status)
	}
}

func (a *Account) Create(ctx *gin.Context) {
	var data dto.CreateUserDto

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := a.services.Account.Create(ctx, &data)

	if status == statuses.OK {
		ctx.JSON(http.StatusCreated, result)
	} else {
		a.handleError(ctx, status)
	}
}

func (a *Account) SignIn(ctx *gin.Context) {
	var body dto.SignInDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := a.services.Account.SignIn(ctx, &body)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		a.handleError(ctx, status)
	}
}

func (a *Account) handleError(ctx *gin.Context, status string) {
	e := &responses.Error{}

	switch status {

	case statuses.Unauthorize:
		e.Error = "Incorrect token."
		ctx.JSON(http.StatusUnauthorized, e)

	case statuses.BadRequest:
		e.Error = "Incorrect data."
		ctx.JSON(http.StatusBadRequest, e)
	
	case statuses.Conflict:
		e.Error = "This username is taken."
		ctx.JSON(http.StatusConflict, e)

	}
}