package entities

import "encoding/json"

type Playlist struct {
	Id        uint64    `json:"id"`
	UserId    uint64    `json:"userId"`
	Title     string    `json:"title"`
	MoviesIds []uint64  `json:"moviesIds"`
}

func (p *Playlist) MarshalBinary() ([]byte, error) {
	return json.Marshal(&p)
}

func (p *Playlist) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &p)
}