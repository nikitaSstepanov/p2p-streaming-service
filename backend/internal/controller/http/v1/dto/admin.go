package dto

import "github.com/nikitaSstepanov/p2p-streaming-service/backend/internal/entity"

const (
	adminRole      = "ADMIN"
	superAdminRole = "SUPER_ADMIN"
)

type AdminDto struct {
	Username string `json:"username"`
	IsSuper  bool   `json:"isSuper"`
}

type AddAdminDto struct {
	Username  string `json:"username"`
	IsSuper   bool   `json:"isSuper"`
}

type RemoveAdminDto struct {
	Username string `json:"username"`
}

func AdminToDto(admin *entity.User) *AdminDto {
	var isSuper bool
	
	if admin.Role == superAdminRole {
		isSuper = true
	} else {
		isSuper = false
	}

	return &AdminDto{
		Username: admin.Username,
		IsSuper:  isSuper,
	}
}