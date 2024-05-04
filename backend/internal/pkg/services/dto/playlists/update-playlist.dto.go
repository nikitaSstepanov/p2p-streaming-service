package dto

type UpdatePlaylistDto struct {
	Title          string   `json:"title"`
	MoviesToAdd    []uint64 `json:"toAdd"`
	MoviesToRemove []uint64 `json:"toRemove"`
}