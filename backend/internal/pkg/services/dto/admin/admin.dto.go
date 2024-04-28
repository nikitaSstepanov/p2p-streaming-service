package dto

type AdminDto struct {
	Username string `json:"username"`
	IsSuper  bool   `json:"isSuper"`
}