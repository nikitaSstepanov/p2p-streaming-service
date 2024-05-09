package dto

type Playlist struct {
	Id     uint64   `json:"id"`
	Title  string   `json:"title"`
	Movies []uint64 `json:"movies"`
}