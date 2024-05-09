package entities

type Comment struct {
	Id      uint64
	MovieId uint64
	UserId  uint64
	Text    string
}