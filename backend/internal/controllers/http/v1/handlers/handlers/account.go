package handlers

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/dto/users"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
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

	result, status, msg := a.services.Account.GetAccount(ctx, header)
	
	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		a.handleError(ctx, status, msg)
	}
}

func (a *Account) Create(ctx *gin.Context) {
	var data dto.CreateUserDto
	
	if err := ctx.ShouldBindJSON(&data); err != nil {
		m := &responses.Message{Message: "Incorrect data."}
		ctx.JSON(http.StatusBadRequest, m)
		return
	}

	result, status, msg := a.services.Account.Create(ctx, &data)

	if status == statuses.OK {
		ctx.JSON(http.StatusCreated, result)
	} else {
		a.handleError(ctx, status, msg)
	}
}

func (a *Account) SignIn(ctx *gin.Context) {
	var body *dto.SignInDto
	
	if err := ctx.ShouldBindJSON(&body); err != nil {
		m := &responses.Message{Message: "Incorrect data."}
		ctx.JSON(http.StatusBadRequest, m)
		return
	}
	
	result, status, msg := a.services.Account.SignIn(ctx, body)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		a.handleError(ctx, status, msg)
	}
}

func (a *Account) handleError(ctx *gin.Context, status string, message string) {
	e := &responses.Error{Error: message}

	switch status {

	case statuses.Unauthorize:
		ctx.JSON(http.StatusUnauthorized, e)

	case statuses.BadRequest:
		ctx.JSON(http.StatusBadRequest, e)
	
	case statuses.Conflict:
		ctx.JSON(http.StatusConflict, e)
	
	case statuses.InternalError:
		ctx.JSON(http.StatusInternalServerError, e)

	}
}