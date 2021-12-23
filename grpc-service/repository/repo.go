package repository

import (
	"context"
	"database/sql"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	erro "github.com/javibauza/final-project/grpc-service/errors"
)

type SQLRepo struct {
	db     *sql.DB
	logger log.Logger
}

type Repository interface {
	Authenticate(ctx context.Context, userName string) (User, error)
	CreateUser(ctx context.Context, user User) error
	UpdateUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, userId string) (User, error)
}

type User struct {
	Id      int
	UserId  string
	PwdHash string
	Name    string
	Age     uint32
	AddInfo sql.NullString
}

func NewRepo(db *sql.DB, logger log.Logger) Repository {
	return &SQLRepo{
		db:     db,
		logger: log.With(logger, "error", "db"),
	}
}

func (repo *SQLRepo) Authenticate(ctx context.Context, userName string) (User, error) {
	logger := log.With(repo.logger, "method", "Authenticate")

	stmt, err := repo.db.Prepare(authenticateSQL)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return User{}, err
	}

	var user User
	err = stmt.QueryRow(userName).Scan(&user.UserId, &user.PwdHash)
	if err != nil {
		if err == sql.ErrNoRows {
			level.Error(logger).Log("err", erro.ErrUserNotFound)
			return User{}, erro.NewErrNotFound()
		}
		level.Error(logger).Log("err", err.Error())
		return User{}, err
	}

	return user, nil
}

func (repo *SQLRepo) CreateUser(ctx context.Context, user User) error {
	logger := log.With(repo.logger, "method", "CreateUser")

	stmt, err := repo.db.Prepare(createSQL)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return err
	}

	_, err = stmt.Exec(user.UserId, user.Name, user.PwdHash, user.Age, user.AddInfo)
	if err != nil {
		level.Error(logger).Log("err", err)
		return err
	}

	return nil
}

func (repo *SQLRepo) UpdateUser(ctx context.Context, user User) error {
	logger := log.With(repo.logger, "method", "UpdateUser")

	args, query := updateSQL(&user)
	stmt, err := repo.db.Prepare(query)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return err
	}

	queryRes, err := stmt.Exec(args...)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return err
	}

	rowCnt, err := queryRes.RowsAffected()
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return err
	}
	if rowCnt == 0 {
		level.Error(logger).Log("err", erro.ErrUserNotFound, "userId", user.UserId)
		return erro.NewErrNotFound()
	}

	return nil
}

func (repo *SQLRepo) GetUser(ctx context.Context, userId string) (User, error) {
	logger := log.With(repo.logger, "method", "GetUser")

	stmt, err := repo.db.Prepare(getSQL)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return User{}, err
	}

	var user User
	err = stmt.QueryRow(userId).Scan(&user.UserId, &user.Name, &user.Age, &user.AddInfo)
	if err != nil {
		if err == sql.ErrNoRows {
			level.Error(logger).Log("err", erro.ErrUserNotFound, "userId", userId)
			return User{}, erro.NewErrNotFound()
		}
		level.Error(logger).Log("err", err.Error())
		return User{}, err
	}

	return user, nil
}
