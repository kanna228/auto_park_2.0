package dto

type CreateDriverRequest struct {
	IIN        string `json:"iin" binding:"required" example:"990101350011"`
	Name       string `json:"name" binding:"required" example:"Ivan"`
	Surname    string `json:"surname" binding:"required" example:"Ivanov"`
	Middlename string `json:"middlename" example:"Ivanovich"`
	Phone      string `json:"phone" example:"+77001234567"`
	Mail       string `json:"mail" example:"driver@mail.com"`
}

type UpdateDriverRequest struct {
	IIN        *string `json:"iin"`
	Name       *string `json:"name"`
	Surname    *string `json:"surname"`
	Middlename *string `json:"middlename"`
	Phone      *string `json:"phone"`
	Mail       *string `json:"mail"`
}
