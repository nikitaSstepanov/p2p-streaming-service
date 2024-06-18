package entities

import "encoding/json"

type Playlist struct {
	Id        uint64    `redis:"id"`
	UserId    uint64    `redis:"userId"`
	Title     string    `redis:"title"`
	MoviesIds []uint64  `redis:"moviesIds"`
}

func (p *Playlist) MarshalBinary() ([]byte, error) {
	return json.Marshal(&p)
}

func (p *Playlist) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}