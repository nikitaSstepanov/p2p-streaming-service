package admin

import (
	"mime/multipart"
	"context"
	"strings"
	"os"
	"io"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"
	e "github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/errors"
	"github.com/google/uuid"
)

const rootAdmin = "admin"

var (
	badAdminReqErr = e.New("You can`t change your role or root admin`s role.", e.BadInput)
	badReqErr      = e.New("Incorrect data.", e.BadInput)
)

type UserStorage interface {
	GetUser(ctx context.Context, id uint64) (*entity.User, *e.Error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, *e.Error)
	GetUsersByRole(ctx context.Context, role string) ([]*entity.User, *e.Error)
	Update(ctx context.Context, user *entity.User) *e.Error
}

type MovieStorage interface {
	GetMovieById(ctx context.Context, id uint64) (*entity.Movie, *e.Error)
	CreateMovie(ctx context.Context, movie *entity.Movie) *e.Error
	UpdateMovie(ctx context.Context, movie *entity.Movie) *e.Error
}

type Admin struct {
	usersStorage  UserStorage
	moviesStorage MovieStorage
}

func New(users UserStorage, movies MovieStorage) *Admin {
	return &Admin{
		usersStorage:  users,
		moviesStorage: movies,
	}
}

func (a *Admin) GetAdmins(ctx context.Context) ([]*entity.User, *e.Error) {
	admins, err := a.usersStorage.GetUsersByRole(ctx, "ADMIN")
	if err != nil {
		return nil, err
	}

	superAdmins, err := a.usersStorage.GetUsersByRole(ctx, "SUPER_ADMIN")
	if err != nil {
		return nil, err
	}

	result := append(admins, superAdmins...)	

	return result, nil
}

func (a *Admin) AddAdmin(ctx context.Context, adminId uint64, username string, isSuper bool) *e.Error {
	user, err := a.usersStorage.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}

	if user.Id == adminId {
		return badAdminReqErr
	}

	if isSuper {
		user.Role = "SUPER_ADMIN"
	} else {
		user.Role = "ADMIN"
	}

	return a.usersStorage.Update(ctx, user)
}

func (a *Admin) RemoveAdmin(ctx context.Context, adminId uint64, username string) *e.Error {
	if username == rootAdmin {
		return badAdminReqErr
	}

	user, err := a.usersStorage.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}

	if user.Id == adminId {
		return badAdminReqErr
	}

	user.Role = "USER"

	return a.usersStorage.Update(ctx, user)
}

func (a *Admin) CreateMovie(ctx context.Context, movie *entity.Movie, files []*multipart.FileHeader) *e.Error {
	paths, err := saveFiles(files)
	if err != nil {
		return err
	}

	movie.Paths = paths

	return a.moviesStorage.CreateMovie(ctx, movie)
}

func (a *Admin) EditMovie(ctx context.Context, updated *entity.Movie, files []*multipart.FileHeader) *e.Error {
	movie, err := a.moviesStorage.GetMovieById(ctx, updated.Id)
	if err != nil {
		return err
	}

	if updated.Name != "" {
		movie.Name = updated.Name
	}

	if len(files) != 0 {
		newPaths, err := saveFiles(files)
		if err != nil {
			return err
		}

		movie.Paths += ";" + newPaths
	}

	return a.moviesStorage.UpdateMovie(ctx, movie)
}

func saveFiles(files []*multipart.FileHeader) (string, *e.Error) {
	if err := checkFiles(files); err != nil {
		return "", err
	}

	paths := []string{}

	for i := 0; i < len(files); i++ {
		file := files[i]
		fileName := uuid.New().String() + ".torrent"

		toSave, err := file.Open()
		if err != nil {
			return "", badReqErr
		}

		local, err := os.OpenFile("files/" + fileName, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return "", badReqErr
		}

		io.Copy(local, toSave)

		paths = append(paths, fileName)
	}

	return strings.Join(paths, ";"), nil
}

func checkFiles(files []*multipart.FileHeader) *e.Error {
	for i := 0; i < len(files); i++ {
		file := files[i]

		if file.Size <= 0 {
			return badReqErr
		}

		parts := strings.Split(file.Filename, ".")

		if parts[len(parts) - 1] != "torrent" {
			return badReqErr
		}
	}

	return nil
}