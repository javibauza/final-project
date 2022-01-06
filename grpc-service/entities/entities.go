package entities

import "database/sql"

type User struct {
	Id       int
	UserId   string
	Password string
	PwdHash  string
	Name     string
	Age      uint32
	AddInfo  sql.NullString
}
