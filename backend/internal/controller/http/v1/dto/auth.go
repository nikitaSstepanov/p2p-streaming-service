package dto

import "github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"

type LoginDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenResult struct {
	Token string `json:"token"`
}

func TokenToDto(tokens *entity.Tokens) *TokenResult {
	return &TokenResult{
		Token: tokens.Access,
	}
}

func (ld *LoginDto) ToEntity() *entity.User {
	return &entity.User{
		Username: ld.Username,
		Password: ld.Password,
	}
}