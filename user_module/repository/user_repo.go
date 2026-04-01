package repository

import "database/sql"

type UserRepo struct {
	DB     *sql.DB
	Schema string
}

func NewUserRepo(db *sql.DB, schema string) *UserRepo {
	return &UserRepo{DB: db, Schema: schema}
}
