package models

import "time"

type Driver struct {
	ID         int64     `json:"id" example:"1"`
	IIN        string    `json:"iin" example:"010203040506"`
	Name       string    `json:"name" example:"Dias"`
	Surname    string    `json:"surname" example:"Abdimanap"`
	Middlename string    `json:"middlename,omitempty" example:"Diasovich"`
	Phone      string    `json:"phone,omitempty" example:"+77001234567"`
	Mail       string    `json:"mail,omitempty" example:"dias@mail.com"`
	CreatedAt  time.Time `json:"created_at" example:"2026-02-18T12:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2026-02-18T12:10:00Z"`
}
