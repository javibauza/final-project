package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/crypto/bcrypt"

	erro "github.com/javibauza/final-project/grpc-service/errors"
	"github.com/javibauza/final-project/grpc-service/repository"
	"github.com/javibauza/final-project/grpc-service/utils"
)

type service struct {
	repository repository.Repository
	logger     log.Logger
}

type AuthRequest struct {
	Name string
	Pwd  string
}

type CreateUserRequest struct {
	Pwd     string
	Name    string
	Age     uint32
	AddInfo string
}

type CreateUserResponse struct {
	UserId string
}

type UpdateUserRequest struct {
	UserId  string
	Name    string
	Pwd     string
	Age     uint32
	AddInfo string
}

type GetUserResponse struct {
	UserId  string
	Name    string
	Age     uint32
	AddInfo string
}

type Service interface {
	Authenticate(ctx context.Context, req AuthRequest) (string, error)
	CreateUser(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error)
	UpdateUser(ctx context.Context, req UpdateUserRequest) error
	GetUser(ctx context.Context, userId string) (GetUserResponse, error)
}

func NewService(rep repository.Repository, logger log.Logger) Service {
	return &service{
		repository: rep,
		logger:     logger,
	}
}

func (s service) Authenticate(ctx context.Context, req AuthRequest) (string, error) {
	logger := log.With(s.logger, "method", "Authenticate")

	if req.Name == "" || req.Pwd == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("name", "password"))
		return "", erro.NewErrInvalidArgument(erro.ErrRequiredFields("name", "password"))
	}

	res, err := s.repository.Authenticate(ctx, req.Name)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(res.PwdHash), []byte(req.Pwd)); err != nil {
		level.Error(logger).Log("err", erro.ErrWrongPassword)
		return "", &erro.ErrPermissionDenied{Err: errors.New(erro.ErrWrongPassword)}
	}

	return res.UserId, nil
}

func (s service) CreateUser(ctx context.Context, req CreateUserRequest) (response CreateUserResponse, err error) {
	logger := log.With(s.logger, "method", "CreateUser")

	if req.Name == "" || req.Pwd == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("name", "password"))
		return CreateUserResponse{}, erro.NewErrInvalidArgument(erro.ErrRequiredFields("name", "password"))
	}

	userId := utils.RandomString(12)
	pwdHash, err := utils.HashPassword(req.Pwd)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return CreateUserResponse{}, err
	}

	err = s.repository.CreateUser(ctx, repository.User{
		UserId:  userId,
		PwdHash: pwdHash,
		Name:    req.Name,
		Age:     req.Age,
		AddInfo: sql.NullString{String: req.AddInfo, Valid: true},
	})
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return CreateUserResponse{}, err
	}

	return CreateUserResponse{UserId: userId}, nil
}

func (s service) UpdateUser(ctx context.Context, req UpdateUserRequest) error {
	logger := log.With(s.logger, "method", "UpdateUser")
	user := repository.User{}

	if req.UserId == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("userId"))
		return erro.NewErrInvalidArgument(erro.ErrRequiredFields("userId"))
	}
	if req.Pwd == "" && req.Age <= 0 && req.AddInfo == "" && req.Name == "" {
		level.Error(logger).Log("err", erro.ErrNoFieldsForUpdate)
		return erro.NewErrInvalidArgument(erro.ErrNoFieldsForUpdate)
	}

	user.UserId = req.UserId
	user.Name = req.Name

	if req.Pwd != "" {
		pwdHash, err := utils.HashPassword(req.Pwd)
		if err != nil {
			level.Error(logger).Log("err", err.Error())
			return err
		}
		user.PwdHash = pwdHash
	}
	if req.Age > 0 {
		user.Age = req.Age
	}
	if req.AddInfo != "" {
		user.AddInfo = sql.NullString{String: req.AddInfo}
	}

	err := s.repository.UpdateUser(ctx, user)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return err
	}

	return nil
}

func (s service) GetUser(ctx context.Context, userId string) (GetUserResponse, error) {
	logger := log.With(s.logger, "method", "GetUser")

	if userId == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("userId"))
		return GetUserResponse{}, erro.NewErrInvalidArgument(erro.ErrRequiredFields("userId"))
	}

	user, err := s.repository.GetUser(ctx, userId)
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return GetUserResponse{}, err
	}

	return GetUserResponse{
		UserId:  user.UserId,
		Name:    user.Name,
		Age:     user.Age,
		AddInfo: user.AddInfo.String,
	}, nil
}
