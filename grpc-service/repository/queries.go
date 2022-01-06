package repository

import (
	"strings"

	"github.com/javibauza/final-project/grpc-service/entities"
)

const authenticateSQL = "SELECT user_id, pwd_hash FROM users WHERE name=?"
const createSQL = "INSERT INTO users (user_id, name, pwd_hash, age, additional_information) VALUES (?, ?, ?, ?, ?)"
const getSQL = "SELECT user_id, name, age, additional_information FROM users WHERE user_id=?"

func updateSQL(user *entities.User) (args []interface{}, query string) {
	query = "UPDATE users"
	queryArgs := " SET "

	if user.PwdHash != "" {
		args = append(args, user.PwdHash)
		queryArgs += " pwd_hash=?,"
	}
	if user.Age > 0 {
		args = append(args, user.Age)
		queryArgs += " age=?,"
	}
	if user.Name != "" {
		args = append(args, user.Name)
		queryArgs += " name=?,"
	}
	if user.AddInfo.String != "" {
		args = append(args, user.AddInfo.String)
		queryArgs += " additional_information=?,"
	}
	queryArgs = strings.TrimSuffix(queryArgs, ",")
	query += queryArgs + " WHERE user_id=?"
	args = append(args, user.UserId)

	return args, query
}
