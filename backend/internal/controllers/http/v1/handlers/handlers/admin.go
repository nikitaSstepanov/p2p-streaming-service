package handlers

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/dto/admin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
	"github.com/gin-gonic/gin"
)

type Admin struct {
	services *services.Services
}

func NewAdmin(services *services.Services) *Admin {
	return &Admin{
		services: services,
	}
}

func (a *Admin) GetAdmins(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	result, status := a.services.Admin.GetAdmins(ctx, header)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		a.handleError(ctx, status)
	}
}

func (a *Admin) AddAdmin(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	var body dto.AddAdminDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := a.services.Admin.AddAdmin(ctx, header, &body)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		a.handleError(ctx, status)
	}
}

func (a *Admin) RemoveAdmin(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	var body dto.RemoveAdminDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := a.services.Admin.RemoveAdmin(ctx, header, &body)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		a.handleError(ctx, status)
	}
}

func (a *Admin) CreateMovie(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	form, err := ctx.MultipartForm()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := a.services.Admin.CreateMovie(ctx, header, form)

	if status == statuses.OK {
		ctx.JSON(http.StatusCreated, result)
	} else {
		a.handleError(ctx, status)
	}
}

func (a *Admin) EditMovie(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	form, err := ctx.MultipartForm()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := a.services.Admin.EditMovie(ctx, header, form)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		a.handleError(ctx, status)
	}
}

func (a *Admin) handleError(ctx *gin.Context, status string) {
	e := &responses.Error{}

	switch status {

	case statuses.Unauthorize:
		e.Error = "Incorrect token."
		ctx.JSON(http.StatusUnauthorized, e)

	case statuses.Forbidden:
		e.Error = "Forbidden."
		ctx.JSON(http.StatusForbidden, e)

	case statuses.NotFound:
		e.Error = "This user wasn`t found.."
		ctx.JSON(http.StatusNoContent, e)

	case statuses.BadRequest:
		e.Error = "Incorrect data."
		ctx.JSON(http.StatusBadRequest, e)

	case statuses.InternalError:
		e.Error = "Something going wrong....."
		ctx.JSON(http.StatusInternalServerError, e)

	}
}