package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/services/dto/admin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
	"github.com/redis/go-redis/v9"
)

type Admin struct {
	Storage *storage.Storage
	Redis   *redis.Client
	Auth    *Auth
}

func NewAdmin(storage *storage.Storage, redis *redis.Client, auth *Auth) *Admin {
	return &Admin{
		Storage: storage,
		Redis:   redis,
		Auth:    auth,
	}
}

func (a *Admin) GetAdmins(ctx *gin.Context) {
	admin := a.checkAccess(ctx, "SUPER_ADMIN")

	if admin.Id == 0 {
		return
	}

	var result []dto.AdminDto

	admins := a.Storage.Users.GetUsersByRole(ctx, "ADMIN")
	superAdmins := a.Storage.Users.GetUsersByRole(ctx, "SUPER_ADMIN")

	for i := 0; i < len(*admins); i++ {
		toAdd := dto.AdminDto{
			Username: (*admins)[i].Username,
			IsSuper:  false,
		}

		result = append(result, toAdd)
	}

	for i := 0; i < len(*superAdmins); i++ {
		toAdd := dto.AdminDto{
			Username: (*superAdmins)[i].Username,
			IsSuper:  true,
		}

		result = append(result, toAdd)
	}

	ctx.JSON(http.StatusOK, result)
}

func (a *Admin) AddAdmin(ctx *gin.Context) {
	admin := a.checkAccess(ctx, "SUPER_ADMIN")

	if admin.Id == 0 {
		return
	}

	var body dto.AddAdminDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	user := a.findUser(ctx, body.Username)

	if user.Id == 0 {
		ctx.JSON(http.StatusNotFound, "User wasn`t found.")
		return
	}

	if admin.Username == body.Username {
		ctx.JSON(http.StatusBadRequest, "It is your username.")
		return
	}

	if body.IsSuper {
		user.Role = "SUPER_ADMIN"
	} else {
		user.Role = "ADMIN"
	}

	a.Storage.Users.Update(ctx, &user)

	ctx.JSON(http.StatusOK, "Role is asigned.")
}

func (a *Admin) RemoveAdmin(ctx *gin.Context) {
	admin := a.checkAccess(ctx, "SUPER_ADMIN")

	if admin.Id == 0 {
		return
	}

	var body dto.RemoveAdminDto

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	user := a.findUser(ctx, body.Username)

	if user.Id == 0 {
		ctx.JSON(http.StatusNotFound, "User wasn`t found.")
		return
	}

	if user.Username == admin.Username {
		ctx.JSON(http.StatusBadRequest, "It is your username.")
		return
	}

	if user.Username == "admin" {
		ctx.JSON(http.StatusBadRequest, "You can`t demote user 'admin'.")
		return
	}

	user.Role = "USER"

	a.Storage.Users.Update(ctx, &user)

	ctx.JSON(http.StatusOK, "Demoted to the user.")
}

func (a *Admin) CreateMovie(ctx *gin.Context) {
	admin := a.checkAccess(ctx, "ADMIN", "SUPER_ADMIN")

	if admin.Id == 0 {
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
		Name:         name[0],
		Paths:        strings.Join(paths, ";"),
		FileVersion:  0,
	}

	a.Storage.Movies.CreateMovie(ctx, movie)

	ctx.JSON(http.StatusCreated, "Created.")
}

func (a *Admin) EditMovie(ctx *gin.Context) {
	admin := a.checkAccess(ctx, "ADMIN", "SUPER_ADMIN")

	if admin.Id == 0 {
		return
	}

	form, err := ctx.MultipartForm()

	if err != nil {
		ctx.JSON(http.StatusBadRequest, "Incorrect data.")
		return
	}

	movieId, isFound := form.Value["movieId"]

	if !isFound {
		ctx.JSON(http.StatusBadRequest, "MovieId is required.")
		return
	}

	movie := a.Storage.Movies.GetMovieById(ctx, movieId[0])

	if movie.Id == 0 {
		ctx.JSON(http.StatusNotFound, "This movie wasn`t found.")
		return
	}

	name, isFound := form.Value["name"]

	if isFound {
		movie.Name = name[0]
	}

	files, isFound := form.File["files"]

	if isFound && len(files) != 0 {
		paths := make([]string, 0)

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

		old := movie.Paths

		new := old + ";" + strings.Join(paths, ";")

		movie.Paths = new
	}

	a.Storage.Movies.UpdateMovie(ctx, movie)

	ctx.JSON(http.StatusOK, "Updated.")
}

func (a *Admin) checkAccess(ctx *gin.Context, roles ...string) *entities.User {
	header := strings.Split(ctx.GetHeader("Authorization"), " ")
	bearer := header[0]
	token := header[1]

	if bearer != "Bearer" {
		ctx.JSON(http.StatusUnauthorized, "Incorrect token.")
		return &entities.User{}
	}

	claims, err := a.Auth.ValidateToken(token)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, "Incorrecct token.")
		return &entities.User{}
	}

	user := a.findUser(ctx, claims.Username)

	if user.Id == 0 {
		ctx.JSON(http.StatusUnauthorized, "Incorrecct token.")
		return &entities.User{}
	}

	found := slices.Contains(roles, user.Role)

	if !found {
		ctx.JSON(http.StatusForbidden, "Forbidden resource.")
		return &entities.User{}
	}

	return &user
}

func (a *Admin) findUser(ctx context.Context, username string) entities.User {
	var user entities.User

	a.Redis.Get(ctx, fmt.Sprintf("users:%s", username)).Scan(&user)

	if user.Id == 0 {
		user = *(a.Storage.Users.GetUser(ctx, username))

		if user.Id == 0 {
			return entities.User{}
		}

		a.Redis.Set(ctx, fmt.Sprintf("users:%s", username), user, 1 * time.Hour)
	}

	return user
}