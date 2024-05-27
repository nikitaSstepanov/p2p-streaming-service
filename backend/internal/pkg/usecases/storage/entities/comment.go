package entities

import "encoding/json"

type Comment struct {
	Id      uint64  `json:"id"`
	MovieId uint64  `json:"movieId"`
	UserId  uint64  `json:"userId"`
	Text    string  `json:"text"`
}

func (c *Comment) MarshalBinary() ([]byte, error) {
	return json.Marshal(&c)
}

func (c *Comment) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &c)
}