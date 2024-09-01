package admin

import (
	"context"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/dto"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/controller/http/v1/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
)

const (
	ok = http.StatusOK
	badReq     = http.StatusBadRequest
	created    = http.StatusCreated
	deleted    = http.StatusNoContent
)

var (
	badReqErr  = e.New("Incorrect data.", e.BadInput)
	addMsg     = responses.NewMessage("Admin was added")
	rmMsg      = responses.NewMessage("Admin was removed.")
	cretaedMsg = responses.NewMessage("New movie created.")
	updatedMsg = responses.NewMessage("Movie updated.")
)

type AdminUseCase interface {
	GetAdmins(ctx context.Context) ([]*entity.User, *e.Error)
	AddAdmin(ctx context.Context, adminId uint64, username string, isSuper bool) *e.Error
	RemoveAdmin(ctx context.Context, adminId uint64, username string) *e.Error
	CreateMovie(ctx context.Context, movie *entity.Movie, files []*multipart.FileHeader) *e.Error
	EditMovie(ctx context.Context, updated *entity.Movie, files []*multipart.FileHeader) *e.Error
}

type Admin struct {
	usecase AdminUseCase
}

func New(uc AdminUseCase) *Admin {
	return &Admin{
		usecase: uc,
	}
}

func (a *Admin) GetAdmins(ctx *gin.Context) {
	admins, err := a.usecase.GetAdmins(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return
	}

	result := make([]dto.AdminDto, 0)

	for i := 0; i < len(admins); i++ {
		result = append(result, *dto.AdminToDto(admins[i]))
	}

	ctx.JSON(ok, result)
}

func (a *Admin) AddAdmin(ctx *gin.Context) {
	adminId := ctx.GetUint64("userId")

	var body dto.AddAdminDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return 
	}

	err := a.usecase.AddAdmin(ctx, adminId, body.Username, body.IsSuper)
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return
	}

	ctx.JSON(created, addMsg)
}

func (a *Admin) RemoveAdmin(ctx *gin.Context) {
	adminId := ctx.GetUint64("userId")

	var body dto.RemoveAdminDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return 
	}

	err := a.usecase.RemoveAdmin(ctx, adminId, body.Username)
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return 
	}

	ctx.JSON(deleted, rmMsg)
}

func (a *Admin) CreateMovie(ctx *gin.Context) {
	form, getFormError := ctx.MultipartForm()
	if getFormError != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return 
	}

	name, isFound := form.Value["name"]

	if !isFound {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return 
	}

	movie := &entity.Movie{
		Name: name[0],	
	}

	files, isFound := form.File["files"]

	if !isFound || len(files) == 0 {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return 
	}

	err := a.usecase.CreateMovie(ctx, movie, files)
	if err != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return 
	}

	ctx.JSON(created, cretaedMsg)
}

func (a *Admin) EditMovie(ctx *gin.Context) {
	form, getFormError := ctx.MultipartForm()
	if getFormError != nil {
		ctx.AbortWithStatusJSON(badReq, badReqErr)
		return 
	}

	var movie entity.Movie

	name, isFound := form.Value["name"]

	if isFound {
		movie.Name = name[0]
	}
 
	files := form.File["files"]

	err := a.usecase.EditMovie(ctx, &movie, files)
	if err != nil {
		ctx.AbortWithStatusJSON(err.ToHttpCode(), err)
		return 
	}

	ctx.JSON(ok, updatedMsg)
}