package dto

type UpdateUserRequest struct {
	Email      *string `json:"email"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	MiddleName *string `json:"middle_name"`
	Password   *string `json:"password"`
	Phone      *string `json:"phone"`
	RoleID     *int64  `json:"role_id"`
	IIN        *string `json:"iin"`
}
