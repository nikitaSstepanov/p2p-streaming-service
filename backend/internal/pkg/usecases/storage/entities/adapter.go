package entities

import "encoding/json"

type Adapter struct {
	Id          uint64  `json:"id"`
	MovieId     uint64  `json:"movieId"`
	Version     uint64  `json:"version"`
	Length      uint64  `json:"length"`
	PieceLength uint64  `json:"pieceLength"`
}

func (a *Adapter) MarshalBinary() ([]byte, error) {
	return json.Marshal(&a)
}

func (a *Adapter) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &a)
}