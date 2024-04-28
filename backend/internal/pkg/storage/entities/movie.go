package entities

import "encoding/json"

type Movie struct {
	Id   		 uint64  `json:"id"`
	Name 		 string	 `json:"name"`
	Paths 		 string  `json:"paths"`
	FileVersion  uint64  `json:"fileVersion"`
}

func (m Movie) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Movie) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &m)
}