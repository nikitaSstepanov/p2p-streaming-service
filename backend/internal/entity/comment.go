package entity

import "encoding/json"

type Comment struct {
	Id      uint64  `redis:"id"`
	MovieId uint64  `redis:"movieId"`
	UserId  uint64  `redis:"userId"`
	Text    string  `redis:"text"`
}

func (c Comment) MarshalBinary() ([]byte, error) {
	return json.Marshal(&c)
}

func (c *Comment) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *Comment) Scan(r row) error {
	return r.Scan(
		&c.Id,
		&c.MovieId,
		&c.UserId,
		&c.Text,
	)
}