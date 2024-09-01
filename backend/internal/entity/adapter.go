package entity

import "encoding/json"

type Adapter struct {
	Id          uint64  `redis:"id"`
	MovieId     uint64  `redis:"movieId"`
	Version     int     `redis:"version"`
	Length      int  `redis:"length"`
	PieceLength int  `redis:"pieceLength"`
}

func (a Adapter) MarshalBinary() ([]byte, error) {
	return json.Marshal(&a)
}

func (a *Adapter) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *Adapter) Scan(r row) error {
	return r.Scan(
		&a.Id,
		&a.MovieId,
		&a.Version,
		&a.Length,
		&a.PieceLength,
	)
}