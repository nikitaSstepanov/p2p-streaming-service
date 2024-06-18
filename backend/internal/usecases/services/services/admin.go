package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/dto/admin"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/responses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/types/statuses"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/redis/go-redis/v9"
)

type Admin struct {
	storage *storage.Storage
	redis   *redis.Client
	account *Account
	auth    *Auth
}

func NewAdmin(storage *storage.Storage, redis *redis.Client, account *Account, auth *Auth) *Admin {
	return &Admin{
		storage: storage,
		redis:   redis,
		account: account,
		auth:    auth,
	}
}

func (a *Admin) GetAdmins(ctx context.Context, header string) (*[]dto.AdminDto, string) {
	_, status := a.checkAccess(ctx, header, "SUPER_ADMIN")

	if status != statuses.OK {
		return nil, status
	}

	var result []dto.AdminDto

	admins, err := a.storage.Users.GetUsersByRole(ctx, "ADMIN")

	if err != nil {
		return nil, statuses.InternalError
	}

	superAdmins, err := a.storage.Users.GetUsersByRole(ctx, "SUPER_ADMIN")

	if err != nil {
		return nil, statuses.InternalError
	}

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

	return &result, statuses.OK
}

func (a *Admin) AddAdmin(ctx context.Context, header string, body *dto.AddAdminDto) (*responses.Message, string) {
	admin, status := a.checkAccess(ctx, header, "SUPER_ADMIN")

	if status != statuses.OK {
		return nil, status
	}

	user, err := a.account.findUser(ctx, body.Username)

	if err != nil {
		return nil, statuses.InternalError
	}

	if user.Id == 0 {
		return nil, statuses.NotFound
	}

	if admin.Username == body.Username {
		return nil, statuses.BadRequest
	}

	if body.IsSuper {
		user.Role = "SUPER_ADMIN"
	} else {
		user.Role = "ADMIN"
	}

	a.storage.Users.Update(ctx, user)

	result := &responses.Message{
		Message: "Role is asigned.",
	}

	return result, statuses.OK
}

func (a *Admin) RemoveAdmin(ctx context.Context, header string, body *dto.RemoveAdminDto) (*responses.Message, string) {
	admin, status := a.checkAccess(ctx, header, "SUPER_ADMIN")

	if status != statuses.OK {
		return nil, status
	}

	user, err := a.account.findUser(ctx, body.Username)

	if err != nil {
		return nil, statuses.InternalError
	}

	if user.Id == 0 {
		return nil, statuses.NotFound
	}

	if user.Username == admin.Username {
		return nil, statuses.BadRequest
	}

	if user.Username == "admin" {
		return nil, statuses.BadRequest
	}

	user.Role = "USER"

	a.storage.Users.Update(ctx, user)

	result := &responses.Message{
		Message: "Demoted to the user.",
	}

	return result, statuses.OK
}

func (a *Admin) CreateMovie(ctx context.Context, header string, form *multipart.Form) (*responses.Message, string) {
	_, status := a.checkAccess(ctx, header, "ADMIN", "SUPER_ADMIN")

	if status != statuses.OK {
		return nil, status
	}

	name, isFound := form.Value["name"]

	if !isFound {
		return nil, statuses.BadRequest
	}

	paths := []string{}

	files, isFound := form.File["files"]

	if isFound && len(files) != 0 {
		for i := 0; i < len(files); i++ {
			file := files[0]

			if file.Size <= 0 {
				return nil, statuses.BadRequest
			}

			parts := strings.Split(file.Filename, ".")

			if parts[len(parts) - 1] != "torrent" {
				return nil, statuses.BadRequest
			}
		}

		for i := 0; i < len(files); i++ {
			file := files[i]
			fileName := uuid.New().String() + ".torrent"

			toSave, err := file.Open()

			if err != nil {
				return nil, statuses.BadRequest
			}

			local, err := os.OpenFile("files/" + fileName, os.O_CREATE|os.O_RDWR, 0644)

			io.Copy(local, toSave)

			if err != nil {
				return nil, statuses.BadRequest
			}

			paths = append(paths, fileName)
		}
	} else {
		return nil, statuses.BadRequest
	}

	movie := &entities.Movie{
		Name:         name[0],
		Paths:        strings.Join(paths, ";"),
		FileVersion:  0,
	}

	a.storage.Movies.CreateMovie(ctx, movie)

	result := &responses.Message{
		Message: "Created.",
	}

	return result, statuses.OK
}

func (a *Admin) EditMovie(ctx context.Context, header string, form *multipart.Form) (*responses.Message, string) {
	_, status := a.checkAccess(ctx, header,  "ADMIN", "SUPER_ADMIN")

	if status != statuses.OK {
		return nil, status
	}

	movieId, isFound := form.Value["movieId"]

	if !isFound {
		return nil, statuses.BadRequest
	}

	movie, err := a.storage.Movies.GetMovieById(ctx, movieId[0])

	if err != nil {
		return nil, statuses.InternalError
	}
	
	if movie.Id == 0 {
		return nil, statuses.NotFound
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
				return nil, statuses.BadRequest
			}

			parts := strings.Split(file.Filename, ".")

			if parts[len(parts) - 1] != "torrent" {
				return nil, statuses.BadRequest
			}
		}

		for i := 0; i < len(files); i++ {
			file := files[i]
			fileName := uuid.New().String() + ".torrent"

			toSave, err := file.Open()

			if err != nil {
				return nil, statuses.BadRequest
			}

			local, err := os.OpenFile("files/" + fileName, os.O_CREATE|os.O_RDWR, 0644)

			io.Copy(local, toSave)

			if err != nil {
				return nil, statuses.BadRequest
			}

			paths = append(paths, fileName)
		}

		old := movie.Paths

		new := old + ";" + strings.Join(paths, ";")

		movie.Paths = new
	}

	a.storage.Movies.UpdateMovie(ctx, &movie)

	result := &responses.Message{
		Message: "Updated.",
	}

	return result, statuses.OK
}

func (a *Admin) checkAccess(ctx context.Context, header string, roles ...string) (*entities.User, string) {
	parts := strings.Split(header, " ")
	bearer := parts[0]
	token := parts[1]

	if bearer != "Bearer" {
		return nil, statuses.Unauthorize
	}

	claims, err := a.auth.ValidateToken(token)
	fmt.Print(err)
	if err != nil {
		return nil, statuses.Unauthorize
	}

	user, err := a.account.findUser(ctx, claims.Username)
	
	if err != nil {
		return nil, statuses.InternalError
	}

	if user.Id == 0 {
		return nil, statuses.Unauthorize
	}

	found := slices.Contains(roles, user.Role)

	if !found {
		return nil, statuses.Forbidden
	}

	return user, statuses.OK
}