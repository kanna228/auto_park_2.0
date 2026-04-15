package dto

type CreateUserRequest struct {
	Email      string  `json:"email" binding:"required"`
	FirstName  string  `json:"first_name" binding:"required"`
	LastName   string  `json:"last_name" binding:"required"`
	MiddleName *string `json:"middle_name"`
	IIN        string  `json:"iin" binding:"required"`
	Phone      *string `json:"phone"`
	RoleID     int64   `json:"role_id" binding:"required"`
}

type CreateUserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	RoleID    int64  `json:"role_id"`
	CreatedAt string `json:"created_at"`
}
