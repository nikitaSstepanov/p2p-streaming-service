package entities

import "encoding/json"

type Adapter struct {
	Id          uint64  `redis:"id"`
	MovieId     uint64  `redis:"movieId"`
	Version     uint64  `redis:"version"`
	Length      uint64  `redis:"length"`
	PieceLength uint64  `redis:"pieceLength"`
}

func (a Adapter) MarshalBinary() ([]byte, error) {
	return json.Marshal(&a)
}

func (a *Adapter) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}