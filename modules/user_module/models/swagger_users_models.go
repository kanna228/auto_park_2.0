package models

// =======================
// SWAGGER DTOs (Users/Auth)
// =======================

// ErrorResponse — единый формат ошибки
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"unauthorized"`
}

// MessageResponse — когда отдаём message
type MessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"ok"`
}

// DeleteUserResponse — ответ удаления
type DeleteUserResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"user deleted successfully"`
	ID      int64  `json:"id" example:"10"`
}

// UserDTO — пользователь для Swagger (публичные поля)
type UserDTO struct {
	ID         int64   `json:"id" example:"1"`
	Email      string  `json:"email" example:"admin@example.com"`
	FirstName  string  `json:"first_name" example:"Dias"`
	LastName   string  `json:"last_name" example:"Kassym"`
	MiddleName *string `json:"middle_name,omitempty" example:"A."`
	IIN        string  `json:"iin" example:"001122334455"`
	Phone      *string `json:"phone,omitempty" example:"+77001234567"`
	RoleID     int64   `json:"role_id" example:"1"`
	LastSeen   *string `json:"last_seen,omitempty" example:"2026-02-18T12:34:56Z"`
	CreatedAt  string  `json:"created_at" example:"2026-02-18T12:34:56Z"`
	UpdatedAt  string  `json:"updated_at" example:"2026-02-18T12:34:56Z"`
}

// UsersListResponse — список пользователей
type UsersListResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    []UserDTO `json:"data"`
}

// UserResponse — один пользователь
type UserResponse struct {
	Success bool    `json:"success" example:"true"`
	Data    UserDTO `json:"data"`
}

// CreateUserResponseSwagger — ответ при создании (у тебя он уже есть в dto, но swagger лучше так)
type CreateUserResponseSwagger struct {
	ID        int64  `json:"id" example:"1"`
	Email     string `json:"email" example:"newuser@example.com"`
	RoleID    int64  `json:"role_id" example:"2"`
	CreatedAt string `json:"created_at" example:"2026-02-18T12:34:56Z"`
}

// CreateUserResponseWrap — обёртка success/data
type CreateUserResponseWrap struct {
	Success bool                      `json:"success" example:"true"`
	Data    CreateUserResponseSwagger `json:"data"`
}

// AuthLoginData — что возвращает login (сделаем универсально)
type AuthLoginData struct {
	Token       string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID      int64  `json:"user_id" example:"1"`
	AccountType string `json:"account_type" example:"driver"`
	DriverID    *int64 `json:"driver_id,omitempty" example:"1"`
	RoleID      int64  `json:"role_id" example:"6"`
	RoleName    string `json:"role_name" example:"driver"`
	Email       string `json:"email" example:"driver@example.com"`
}

// AuthLoginResponse — обёртка
type AuthLoginResponse struct {
	Success bool          `json:"success" example:"true"`
	Data    AuthLoginData `json:"data"`
}
