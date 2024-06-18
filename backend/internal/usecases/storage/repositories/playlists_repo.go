package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/usecases/storage/entities"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/pkg/postgresql"
)

const (
	playlistsTable = "playlists"
	playlistsMoviesTable = "movies_playlists"
)

type Playlists struct {
	db postgresql.Client
}

func NewPlaylists(db postgresql.Client) *Playlists {
	return &Playlists{
		db: db,
	}
}

func (p *Playlists) GetPlaylists(ctx context.Context, userId uint64, limit string, offset string) (*[]entities.Playlist, error) {
	var playlists []entities.Playlist

	var playlist entities.Playlist
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE userId = %d LIMIT %s OFFSET %s;", playlistsTable, userId, limit, offset)

	rows, err := p.db.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&playlist.Id, &playlist.UserId, &playlist.Title)

		if err != nil {
			return nil, err
		}

		playlists = append(playlists, playlist)
	}

	return &playlists, nil
}

func (p *Playlists) GetPlaylist(ctx context.Context, id string) (*entities.Playlist, error) {
	var playlist entities.Playlist

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %s;", playlistsTable, id)

	row := p.db.QueryRow(ctx, query)

	err := row.Scan(&playlist.Id, &playlist.UserId, &playlist.Title)

	if err != nil {
		return nil, err
	}

	var movies []uint64

	var movie entities.PlaylistMovies

	query = fmt.Sprintf("SELECT * FROM %s WHERE playlistId = %s;", playlistsMoviesTable, id)

	rows, err := p.db.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&movie.PlaylistId, &movie.MovieId)

		if err != nil {
			return nil, err
		}

		movies = append(movies, movie.MovieId)
	}

	playlist.MoviesIds = movies

	return &playlist, nil
}

func (p *Playlists) CreatePlaylist(ctx context.Context, playlist *entities.Playlist) error {
	query := fmt.Sprintf("INSERT INTO %s (userId, title) VALUES ($1, $2);", playlistsTable)

	tx, err := p.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, playlist.UserId, playlist.Title)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (p *Playlists) UpdatePlaylist(ctx context.Context, playlist entities.Playlist) error {
	query := fmt.Sprintf("UPDATE %s SET title = $1 WHERE id = $2;", playlistsTable)

	tx, err := p.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, playlist.Title, playlist.Id)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (p *Playlists) AddMovie(ctx context.Context, playlistId uint64, movieId uint64) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE playlistId = %d AND movieId = %d ;", playlistsMoviesTable, playlistId, movieId)

	row := p.db.QueryRow(ctx, query)

	var movie entities.PlaylistMovies

	err := row.Scan(&movie.PlaylistId, &movie.MovieId)

	if err != nil {
		return err
	}
	
	if movie.MovieId != 0 {
		return nil
	}

	query = fmt.Sprintf("INSERT INTO %s (playlistId, movieId) VALUES ($1, $2) ON CONFLICT DO NOTHING;", playlistsMoviesTable)

	tx, err := p.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, playlistId, movieId)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (p *Playlists) RemoveMovie(ctx context.Context, playlistId uint64, movieId uint64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE playlistId = $1 AND movieId = $2;", playlistsMoviesTable)

	tx, err := p.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, playlistId, movieId)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (p *Playlists) DeletePlaylist(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1;", playlistsTable)

	tx, err := p.db.Begin(ctx)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, id)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)

	if err != nil {
		return err
	}

	return nil
}