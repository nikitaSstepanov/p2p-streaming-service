package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
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

func (m *Movies) GetAllMovies(ctx context.Context, limit string, offset string) *[]entities.Movie {
	var movies []entities.Movie

	var movie entities.Movie

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %s OFFSET %s;", moviesTable, limit, offset)

	rows, _ := m.db.Query(ctx, query)

	for rows.Next() {
		rows.Scan(&movie.Id, &movie.Name, &movie.Path)

		movies = append(movies, movie)
	}

	return &movies
}

func (m *Movies) GetMovieById(ctx context.Context, id string) *entities.Movie {
	var movie entities.Movie
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %s;", moviesTable, id)

	row := m.db.QueryRow(ctx, query)

	row.Scan(&movie.Id, &movie.Name, &movie.Path)

	return &movie
}
