package dto

type CommentDto struct {
	Id     uint64 `json:"id"`
	UserId uint64 `json:"userId"`
	Text   string `json:"text"`
}