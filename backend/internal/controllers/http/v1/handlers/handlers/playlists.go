package handlers

import (
	"net/http"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/dto/playlists"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/services"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
	"github.com/gin-gonic/gin"
)

type Playlists struct {
	services *services.Services
}

func NewPlaylists(services *services.Services) *Playlists {
	return &Playlists{
		services: services,
	}
}

func (p *Playlists) GetPlaylists(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	limit := ctx.DefaultQuery("limit", "5")

	offset := ctx.DefaultQuery("offset", "0")

	result, status := p.services.Playlists.GetPlaylists(ctx, header, limit, offset)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		p.handleError(ctx, status)
	}
}

func (p *Playlists) GetPlaylistById(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	id := ctx.Param("id")

	result, status := p.services.Playlists.GetPlaylistById(ctx, header, id)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		p.handleError(ctx, status)
	}
}

func (p *Playlists) CreatePlaylist(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	var body dto.CreatePlaylistDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := p.services.Playlists.CreatePlaylist(ctx, header, &body)

	if status == statuses.OK {
		ctx.JSON(http.StatusCreated, result)
	} else {
		p.handleError(ctx, status)
	}
}

func (p *Playlists) EditPlaylist(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	id := ctx.Param("id")
	
	var body dto.UpdatePlaylistDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	result, status := p.services.Playlists.EditPlaylist(ctx, header, id, &body)

	if status == statuses.OK {
		ctx.JSON(http.StatusOK, result)
	} else {
		p.handleError(ctx, status)
	}
}

func (p *Playlists) DeletePlaylist(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")

	id := ctx.Param("id")

	result, status := p.services.Playlists.DeletePlaylist(ctx, header, id)

	if status == statuses.OK {
		ctx.JSON(http.StatusNoContent, result)
	} else {
		p.handleError(ctx, status)
	}
}

func (p *Playlists) handleError(ctx *gin.Context, status string) {
	e := &responses.Error{}

	switch status {

	case statuses.Unauthorize:
		e.Error = "Incorrect token."
		ctx.JSON(http.StatusUnauthorized, e)

	case statuses.NotFound:
		e.Error = "This playlist wasn`t found."
		ctx.JSON(http.StatusNotFound, e)

	case statuses.Forbidden:
		e.Error = "Forbidden."
		ctx.JSON(http.StatusForbidden, e)

	case statuses.InternalError:
		e.Error = "Something going wrong....."
		ctx.JSON(http.StatusInternalServerError, e)

	}
}