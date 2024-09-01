package dto

import "github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"

type AccountDto struct {
	Username string `json:"username"`
}

type CreateUserDto struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

func AccountToDto(user *entity.User) *AccountDto {
	return &AccountDto{
		Username: user.Username,
	}
}

func (c *CreateUserDto) ToEntity() *entity.User {
	return &entity.User{
		Username: c.Username,
		Password: c.Password,
	}
}