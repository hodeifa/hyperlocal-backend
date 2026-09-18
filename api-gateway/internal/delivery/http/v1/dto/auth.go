package dto

type RegisterRequestDTO struct {
	PhoneNumber string `json:"phone_number" binding:"required,e164"`
}