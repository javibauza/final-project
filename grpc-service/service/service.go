package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/crypto/bcrypt"

	"github.com/javibauza/final-project/grpc-service/entities"
	erro "github.com/javibauza/final-project/grpc-service/errors"
	"github.com/javibauza/final-project/grpc-service/utils"
)

type service struct {
	repository Repository
	logger     log.Logger
}

type Repository interface {
	Authenticate(ctx context.Context, userName string) (entities.User, error)
	CreateUser(ctx context.Context, user entities.User) error
	UpdateUser(ctx context.Context, user entities.User) error
	GetUser(ctx context.Context, userId string) (entities.User, error)
}

func NewService(rep Repository, logger log.Logger) *service {
	return &service{
		repository: rep,
		logger:     logger,
	}
}

func (s service) Authenticate(ctx context.Context, req entities.User) (string, error) {
	logger := log.With(s.logger, "method", "Authenticate")

	if req.Name == "" || req.Password == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("name", "password"))
		return "", erro.NewErrInvalidArgument(erro.ErrRequiredFields("name", "password"))
	}

	res, err := s.repository.Authenticate(ctx, req.Name)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(res.PwdHash), []byte(req.Password)); err != nil {
		level.Error(logger).Log("err", erro.ErrWrongPassword)
		return "", &erro.ErrPermissionDenied{Err: errors.New(erro.ErrWrongPassword)}
	}

	return res.UserId, nil
}

func (s service) CreateUser(ctx context.Context, req entities.User) (response entities.User, err error) {
	logger := log.With(s.logger, "method", "CreateUser")

	if req.Name == "" || req.Password == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("name", "password"))
		return entities.User{}, erro.NewErrInvalidArgument(erro.ErrRequiredFields("name", "password"))
	}

	userId := utils.RandomString(12)
	pwdHash, err := utils.HashPassword(req.Password)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return entities.User{}, err
	}

	err = s.repository.CreateUser(ctx, entities.User{
		UserId:  userId,
		PwdHash: pwdHash,
		Name:    req.Name,
		Age:     req.Age,
		AddInfo: req.AddInfo,
	})
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return entities.User{}, err
	}

	return entities.User{UserId: userId}, nil
}

func (s service) UpdateUser(ctx context.Context, req entities.User) error {
	logger := log.With(s.logger, "method", "UpdateUser")
	user := entities.User{}

	if req.UserId == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("userId"))
		return erro.NewErrInvalidArgument(erro.ErrRequiredFields("userId"))
	}
	if req.Password == "" && req.Age <= 0 && req.AddInfo.String == "" && req.Name == "" {
		level.Error(logger).Log("err", erro.ErrNoFieldsForUpdate)
		return erro.NewErrInvalidArgument(erro.ErrNoFieldsForUpdate)
	}

	user.UserId = req.UserId
	user.Name = req.Name

	if req.Password != "" {
		pwdHash, err := utils.HashPassword(req.Password)
		if err != nil {
			level.Error(logger).Log("err", err.Error())
			return err
		}
		user.PwdHash = pwdHash
	}
	if req.Age > 0 {
		user.Age = req.Age
	}
	if req.AddInfo.String != "" {
		user.AddInfo = sql.NullString{String: req.AddInfo.String}
	}

	err := s.repository.UpdateUser(ctx, user)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return err
	}

	return nil
}

func (s service) GetUser(ctx context.Context, userId string) (entities.User, error) {
	logger := log.With(s.logger, "method", "GetUser")

	if userId == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("userId"))
		return entities.User{}, erro.NewErrInvalidArgument(erro.ErrRequiredFields("userId"))
	}

	user, err := s.repository.GetUser(ctx, userId)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return entities.User{}, err
	}

	return entities.User{
		UserId:  user.UserId,
		Name:    user.Name,
		Age:     user.Age,
		AddInfo: user.AddInfo,
	}, nil
}
