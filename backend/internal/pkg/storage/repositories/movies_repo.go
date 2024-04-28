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
		rows.Scan(&movie.Id, &movie.Name, &movie.Paths, &movie.FileVersion)

		movies = append(movies, movie)
	}

	return &movies
}

func (m *Movies) GetMovieById(ctx context.Context, id string) *entities.Movie {
	var movie entities.Movie
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %s;", moviesTable, id)

	row := m.db.QueryRow(ctx, query)

	row.Scan(&movie.Id, &movie.Name, &movie.Paths, &movie.FileVersion)

	return &movie
}

func (m *Movies) CreateMovie(ctx context.Context, movie *entities.Movie) {
	query := fmt.Sprintf("INSERT INTO %s (name, paths, fileVersion) VALUES ('%s', '%s', '%d');", moviesTable, movie.Name, movie.Paths, movie.FileVersion)

	m.db.QueryRow(ctx, query)
}

func (m *Movies) UpdateMovie(ctx context.Context, movie *entities.Movie) {
	fmt.Println(movie)
	query := fmt.Sprintf("UPDATE %s SET fileVersion = %d, paths = '%s' WHERE id = %d;", moviesTable, movie.FileVersion, movie.Paths, movie.Id)

	m.db.QueryRow(ctx, query)
}