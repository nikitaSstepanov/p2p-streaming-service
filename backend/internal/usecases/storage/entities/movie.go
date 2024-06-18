package entities

import "encoding/json"

type Movie struct {
	Id   		 uint64  `redis:"id"`
	Name 		 string	 `redis:"name"`
	Paths 		 string  `redis:"paths"`
	FileVersion  uint64  `redis:"fileVersion"`
}

func (m Movie) MarshalBinary() ([]byte, error) {
	return json.Marshal(&m)
}

func (m *Movie) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}