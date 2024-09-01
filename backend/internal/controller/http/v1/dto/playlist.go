package dto

import "github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"

type CreatePlaylistDto struct {
	Title string `json:"title"`
}

type PlaylistForList struct {
	Id     uint64   `json:"id"`
	Title  string   `json:"title"`
}

type Playlist struct {
	Id     uint64   `json:"id"`
	Title  string   `json:"title"`
	Movies []uint64 `json:"movies"`
}

type UpdatePlaylistDto struct {
	Title          string   `json:"title"`
	MoviesToAdd    []uint64 `json:"toAdd"`
	MoviesToRemove []uint64 `json:"toRemove"`
}

func PlaylistToDto(playlit *entity.Playlist) Playlist {
	return Playlist{
		Id: playlit.Id,
		Title: playlit.Title,
		Movies: playlit.MoviesIds,
	}
}

func PlaylistToList(Playlist *entity.Playlist) PlaylistForList {
	return PlaylistForList{
		Id: Playlist.Id,
		Title: Playlist.Title,
	}
}