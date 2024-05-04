package entities

type Playlist struct {
	Id        uint64
	UserId    uint64   
	Title     string
	MoviesIds []uint64
}