package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	IIN      string `json:"iin" binding:"required"`
}
