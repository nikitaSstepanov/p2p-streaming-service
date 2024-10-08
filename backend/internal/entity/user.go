package entity

import "encoding/json"

type User struct {
	Id       uint64  `redis:"id"`
	Username string  `redis:"username"`
	Password string  `redis:"password"`
	Role     string  `redis:"role"`
}

func (u User) MarshalBinary() ([]byte, error) {
	return json.Marshal(&u)
}
 
func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u *User) Scan(r row) error {
	return r.Scan(
		&u.Id, 
		&u.Username, 
		&u.Password, 
		&u.Role,
	)
}