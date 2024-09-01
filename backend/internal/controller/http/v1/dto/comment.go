package dto

import "github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"

type CommentDto struct {
	Id     uint64 `json:"id"`
	UserId uint64 `json:"userId"`
	Text   string `json:"text"`
}

type CreateCommentDto struct {
	Text string `json:"text"`
}

type UpdateCommentDto struct {
	Text string `json:"text"`
}

func CommentToDto(comment *entity.Comment) *CommentDto {
	return &CommentDto{
		Id: comment.Id,
		UserId: comment.UserId,
		Text: comment.Text,
	}
}