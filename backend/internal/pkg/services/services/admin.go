package services

import (
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
)

var (
	available = []string{ "ADMIN", "SUPER_ADMIN" }
)

type Admin struct {
	Storage *storage.Storage
	Auth *Auth
}

func NewAdmin(storage *storage.Storage, auth *Auth) *Admin {
	return &Admin{
		Storage: storage,
		Auth: auth,
	}
}

func (a *Admin) CreateMovie(ctx *gin.Context) {
	header := strings.Split(ctx.GetHeader("Authorization"), " ")
	bearer := header[0]
	token := header[1]

	if bearer != "Bearer" {
		ctx.JSON(http.StatusUnauthorized, "Incorrect token.")
		return
	}

	claims, err := a.Auth.ValidateToken(token)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "Incorrecct token.")
		return
	}

	user := a.Storage.Users.GetUser(ctx, claims.Username)

	if user.Id == 0 {
		ctx.JSON(http.StatusUnauthorized, "Incorrecct token.")
		return
	}

	found := slices.Contains(available, user.Role)

	if !found {
		ctx.JSON(http.StatusForbidden, "Forbidden resource.")
		return
	}

	form, err := ctx.MultipartForm()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	name, isFound := form.Value["name"]

	if !isFound {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	paths := []string{}

	files, isFound := form.File["files"]

	if isFound && len(files) != 0 {
		for i := 0; i < len(files); i++ {
			file := files[0]

			if file.Size <= 0 {
				ctx.JSON(http.StatusBadRequest, "Incorrect data.")
				return
			}

			parts := strings.Split(file.Filename, ".")

			if parts[len(parts) - 1] != "torrent" {
				ctx.JSON(http.StatusBadRequest, "Incorrect data.")
				return
			}
		}

		for i := 0; i < len(files); i++ {
			file := files[i]
			fileName := uuid.New().String() + ".torrent"

			toSave, err := file.Open()

			if err != nil {
				ctx.JSON(http.StatusBadRequest, "Incorrect data.")
				return
			}

			local, err := os.OpenFile("files/" + fileName, os.O_CREATE|os.O_RDWR, 0644)

			io.Copy(local, toSave)

			if err != nil {
				ctx.JSON(http.StatusBadRequest, "Incorrect data.")
				return
			}

			paths = append(paths, fileName)
		}
	}

	movie := &entities.Movie{
		Name: name[0],
		Paths: strings.Join(paths, ";"),
		FileVersion: 0,
	}

	a.Storage.Movies.CreateMovie(ctx, movie)

	ctx.JSON(http.StatusCreated, "Created.")
}