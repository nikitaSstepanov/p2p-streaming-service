package dto

type AddAdminDto struct {
	Username  string `json:"username"`
	IsSuper   bool   `json:"isSuper"`
}