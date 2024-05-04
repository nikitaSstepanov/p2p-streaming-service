package repositories

import (
	"context"
	"fmt"

	"github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/pkg/storage/entities"
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

func (p *Playlists) GetPlaylists(ctx context.Context, userId uint64, limit string, offset string) *[]entities.Playlist {
	var playlists []entities.Playlist

	var playlist entities.Playlist
	
	query := fmt.Sprintf("SELECT * FROM %s WHERE userId = %d LIMIT %s OFFSET %s;", playlistsTable, userId, limit, offset)

	rows, _ := p.db.Query(ctx, query)

	for rows.Next() {
		rows.Scan(&playlist.Id, &playlist.UserId, &playlist.Title)

		playlists = append(playlists, playlist)
	}

	return &playlists
}

func (p *Playlists) GetPlaylist(ctx context.Context, id string) *entities.Playlist {
	var playlist entities.Playlist

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = %s;", playlistsTable, id)

	row := p.db.QueryRow(ctx, query)

	row.Scan(&playlist.Id, &playlist.UserId, &playlist.Title)

	var movies []uint64

	var movie entities.PlaylistMovies

	query = fmt.Sprintf("SELECT * FROM %s WHERE playlistId = %s;", playlistsMoviesTable, id)

	rows, _ := p.db.Query(ctx, query)

	for rows.Next() {
		rows.Scan(&movie.PlaylistId, &movie.MovieId)

		movies = append(movies, movie.MovieId)
	}

	playlist.MoviesIds = movies

	return &playlist
}

func (p *Playlists) CreatePlaylist(ctx context.Context, playlist entities.Playlist) {
	query := fmt.Sprintf("INSERT INTO %s (userId, title) VALUES (%d, '%s');", playlistsTable, playlist.UserId, playlist.Title)

	p.db.QueryRow(ctx, query)
}

func (p *Playlists) UpdatePlaylist(ctx context.Context, playlist entities.Playlist) {
	query := fmt.Sprintf("UPDATE %s SET title = '%s' WHERE id = %d;", playlistsTable, playlist.Title, playlist.Id)

	p.db.QueryRow(ctx, query)
}

func (p *Playlists) AddMovie(ctx context.Context, playlistId uint64, movieId uint64) {
	query := fmt.Sprintf("INSERT INTO %s (playlistId, movieId) VALUES (%d, %d) ON CONFLICT DO NOTHING;", playlistsMoviesTable, playlistId, movieId)

	p.db.QueryRow(ctx, query)
}

func (p *Playlists) RemoveMovie(ctx context.Context, playlistId uint64, movieId uint64) {
	query := fmt.Sprintf("DELETE FROM %s WHERE playlistId = %d AND movieId = %d;", playlistsMoviesTable, playlistId, movieId)

	p.db.QueryRow(ctx, query)
}

func (p *Playlists) DeletePlaylist(ctx context.Context, id string) {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = %s;", playlistsTable, id)

	p.db.QueryRow(ctx, query)
}