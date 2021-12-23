package service

import (
	"context"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	erro "github.com/javibauza/final-project/gbfp-httpservice/errors"
	"github.com/javibauza/final-project/gbfp-httpservice/repository"
)

type Service interface {
	Authenticate(ctx context.Context, request AuthRequest) (AuthResponse, error)
	CreateUser(ctx context.Context, request CreateUserRequest) (CreateUserResponse, error)
	UpdateUser(ctx context.Context, request UpdateUserRequest) error
	GetUser(ctx context.Context, userId string) (GetUserResponse, error)
}

type service struct {
	repository repository.UserRepository
	logger     log.Logger
}

type AuthRequest struct {
	Name string
	Pwd  string
}

type AuthResponse struct {
	UserId string
}

type CreateUserRequest struct {
	Name    string
	Pwd     string
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

func NewService(rep repository.UserRepository, logger log.Logger) Service {
	return &service{
		repository: rep,
		logger:     logger,
	}
}

func (s service) Authenticate(ctx context.Context, request AuthRequest) (AuthResponse, error) {
	logger := log.With(s.logger, "method", "Authenticate")

	if request.Name == "" || request.Pwd == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("name", "password"))
		return AuthResponse{}, erro.NewErrBadRequest(erro.ErrRequiredFields("name", "password"))
	}

	res, err := s.repository.Authenticate(ctx, repository.User{
		Name:     request.Name,
		Password: request.Pwd,
	})
	if err != nil {
		level.Error(logger).Log("err", err.Error())
		return AuthResponse{}, err
	}

	return AuthResponse{UserId: res.UserId}, nil

}

func (s service) CreateUser(ctx context.Context, request CreateUserRequest) (CreateUserResponse, error) {
	logger := log.With(s.logger, "method", "CreateUser")

	if request.Name == "" || request.Pwd == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("name", "password"))
		return CreateUserResponse{}, erro.NewErrBadRequest(erro.ErrRequiredFields("name", "password"))
	}

	userId, err := s.repository.CreateUser(ctx, repository.User{
		Name:     request.Name,
		Password: request.Pwd,
		Age:      request.Age,
		AddInfo:  request.AddInfo,
	})
	if err != nil {
		level.Error(logger).Log("err", err)
		return CreateUserResponse{}, err
	}

	return CreateUserResponse{UserId: userId}, nil
}

func (s service) UpdateUser(ctx context.Context, request UpdateUserRequest) error {
	logger := log.With(s.logger, "method", "UpdateUser")

	if request.UserId == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("userId"))
		return erro.NewErrBadRequest(erro.ErrRequiredFields("userId"))
	}
	if request.Pwd == "" && request.Age <= 0 && request.AddInfo == "" && request.Name == "" {
		return erro.NewErrBadRequest(erro.ErrNoFieldsForUpdate)
	}

	err := s.repository.UpdateUser(ctx, repository.User{
		UserId:   request.UserId,
		Name:     request.Name,
		Password: request.Pwd,
		Age:      request.Age,
		AddInfo:  request.AddInfo,
	})
	if err != nil {
		level.Error(logger).Log("err", err)
		return err
	}

	return nil
}

func (s service) GetUser(ctx context.Context, userId string) (GetUserResponse, error) {
	logger := log.With(s.logger, "method", "UpdateUser")

	if userId == "" {
		level.Error(logger).Log("err", erro.ErrRequiredFields("userId"))
		return GetUserResponse{}, erro.NewErrBadRequest(erro.ErrRequiredFields("userId"))
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
		AddInfo: user.AddInfo,
	}, nil
}
