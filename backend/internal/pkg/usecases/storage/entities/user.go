package entities

import "encoding/json"

type User struct {
	Id       uint64  `json:"id"`
	Username string  `json:"username"`
	Password string  `json:"password"`
	Role     string  `json:"role"`
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(&u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &u)
}
