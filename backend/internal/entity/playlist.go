package entity

import "encoding/json"

type Playlist struct {
	Id        uint64    `redis:"id"`
	UserId    uint64    `redis:"userId"`
	Title     string    `redis:"title"`
	MoviesIds []uint64  `redis:"moviesIds"`
}

type PlaylistMovies struct {
	PlaylistId uint64
	MovieId    uint64
}

func (p *Playlist) MarshalBinary() ([]byte, error) {
	return json.Marshal(&p)
}

func (p *Playlist) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *Playlist) Scan(r row) error {
	return r.Scan(
		&p.Id, 
		&p.UserId, 
		&p.Title,
	)
}

func (pm *PlaylistMovies) Scan(r row) error {
	return r.Scan(
		&pm.PlaylistId, 
		&pm.MovieId,
	)
}