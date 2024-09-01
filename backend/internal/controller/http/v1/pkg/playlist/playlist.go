package playlist

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/dto"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

const (
	ok = http.StatusOK
	badReq = http.StatusBadRequest
	created = http.StatusCreated
	deleted = http.StatusNoContent
)

var (
	badReqErr  = e.New("Incorrect data.", e.BadInput)

	createdMsg = responses.NewMessage("Created.")
	updatedMsg = responses.NewMessage("Updated.")
	deletedMsg = responses.NewMessage("Deleted.")
)

type PlaylistUseCase interface {
	GetPlaylists(ctx context.Context, id uint64, limit int, offset int) ([]*entity.Playlist, *e.Error)
	GetPlaylistById(ctx context.Context, userId uint64, id uint64) (*entity.Playlist, *e.Error)
	CreatePlaylist(ctx context.Context, playlist *entity.Playlist) *e.Error
	EditPlaylist(ctx context.Context, updated *entity.Playlist, toAdd []uint64, toRemove []uint64) *e.Error
	DeletePlaylist(ctx context.Context, userId uint64, id uint64) *e.Error
}

type Playlist struct {
	usecase PlaylistUseCase
}

func New(usecase PlaylistUseCase) *Playlist {
	return &Playlist{
		usecase,
	}
}

func (p *Playlist) GetPlaylists(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "5"))
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	playlists, playlistErr := p.usecase.GetPlaylists(ctx, userId, limit, offset)
	if playlistErr != nil {
		ctx.AbortWithStatusJSON(playlistErr.ToHttpCode(), playlistErr)
		return
	}

	result := make([]dto.PlaylistForList, 0)

	for i := 0; i < len(playlists); i++ {
		result = append(result, dto.PlaylistToList(playlists[i]))
	}

	ctx.JSON(ok, result)
}

func (p *Playlist) GetPlaylistById(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	playlist, playlistErr := p.usecase.GetPlaylistById(ctx, userId, id)
	if playlistErr != nil {
		ctx.AbortWithStatusJSON(playlistErr.ToHttpCode(), playlistErr)
		return
	}

	ctx.JSON(ok, dto.PlaylistToDto(playlist))
}

func (p *Playlist) CreatePlaylist(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")
	
	var body dto.CreatePlaylistDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	playlist := &entity.Playlist{
		UserId: userId,
		Title: body.Title,
	}

	playlistErr := p.usecase.CreatePlaylist(ctx, playlist)
	if playlistErr != nil {
		ctx.AbortWithStatusJSON(playlistErr.ToHttpCode(), playlistErr)
		return
	}

	ctx.JSON(created, createdMsg)
}

func (p *Playlist) EditPlaylist(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}
	
	var body dto.UpdatePlaylistDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(badReq, badReqErr)
		return
	}

	playlist := &entity.Playlist{
		Id: id,
		UserId: userId,
		Title: body.Title,
	}

	playlistErr := p.usecase.EditPlaylist(ctx, playlist, body.MoviesToAdd, body.MoviesToRemove)
	if playlistErr != nil {
		ctx.AbortWithStatusJSON(playlistErr.ToHttpCode(), playlistErr)
		return
	}

	ctx.JSON(ok, updatedMsg)
}

func (p *Playlist) DeletePlaylist(ctx *gin.Context) {
	userId := ctx.GetUint64("userId")

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return
	}

	playlistErr := p.usecase.DeletePlaylist(ctx, userId, id)
	if playlistErr != nil {
		ctx.AbortWithStatusJSON(playlistErr.ToHttpCode(), playlistErr)
		return
	}

	ctx.JSON(deleted, deletedMsg)
}