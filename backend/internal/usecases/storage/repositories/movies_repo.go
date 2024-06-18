package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/postgresql"
)

const (
	moviesTable = "movies"
)

type Movies struct {
	db postgresql.Client
}

func NewMovies(db postgresql.Client) *Movies {
	return &Movies{
		db: db,
	}
}

func (m *Movies) GetAllMovies(ctx context.Context, limit string, offset string) (*[]entities.Movie, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %s OFFSET %s;", moviesTable, limit, offset)

	rows, err := m.db.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	var movies []entities.Movie

	var movie entities.Movie

	for rows.Next() {
		err = rows.Scan(&movie.Id, &movie.Name, &movie.Paths, &movie.FileVersion)

		if err != nil {
			return nil, err
		}

		movies = append(movies, movie)
	}

	return &movies, nil
}

func (m *Movies) GetMovieById(ctx context.Context, id string) (entities.Movie, error) {
	var movie entities.Movie
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %s;", moviesTable, id)

	row := m.db.QueryRow(ctx, query)

	err := row.Scan(&movie.Id, &movie.Name, &movie.Paths, &movie.FileVersion)

	if err != nil {
		return entities.Movie{}, err
	}

	return movie, nil
}

func (m *Movies) CreateMovie(ctx context.Context, movie *entities.Movie) error {
	query := fmt.Sprintf("INSERT INTO %s (name, paths, fileVersion) VALUES ($1, $2, $3);", moviesTable)

	tx, err := m.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query,  movie.Name, movie.Paths, movie.FileVersion)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (m *Movies) UpdateMovie(ctx context.Context, movie *entities.Movie) error {
	query := fmt.Sprintf("UPDATE %s SET fileVersion = $1, paths = $2, name = $3 WHERE id = $4;", moviesTable)

	tx, err := m.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, movie.FileVersion, movie.Paths, movie.Name, movie.Id)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}