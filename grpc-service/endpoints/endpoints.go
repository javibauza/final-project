package endpoints

import (
	"context"
	"database/sql"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	//"fmt"

	"github.com/go-kit/kit/endpoint"

	"github.com/javibauza/final-project/grpc-service/entities"
	erro "github.com/javibauza/final-project/grpc-service/errors"
)

type Service interface {
	Authenticate(ctx context.Context, req entities.User) (string, error)
	CreateUser(ctx context.Context, req entities.User) (entities.User, error)
	UpdateUser(ctx context.Context, req entities.User) error
	GetUser(ctx context.Context, userId string) (entities.User, error)
}

type Endpoints struct {
	Authenticate endpoint.Endpoint
	CreateUser   endpoint.Endpoint
	UpdateUser   endpoint.Endpoint
	GetUser      endpoint.Endpoint
}

type AuthRequest struct {
	Pwd  string
	Name string
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

type GetUserRequest struct {
	UserId string
}
type GetUserResponse struct {
	UserId  string
	Name    string
	Age     uint32
	AddInfo string
}

func MakeEndpoints(s Service, logger log.Logger) Endpoints {
	return Endpoints{
		Authenticate: makeAuthEndpoint(s, logger),
		CreateUser:   makeCreateUserEndpoint(s, logger),
		UpdateUser:   makeUpdateUserEndpoint(s, logger),
		GetUser:      makeGetUserEndpoint(s, logger),
	}
}

func makeAuthEndpoint(s Service, logger log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		logger := log.With(logger, "method", "makeAuthEndpoint")
		req, ok := request.(AuthRequest)
		if !ok {
			level.Error(logger).Log("error", erro.ErrInvalidRequestType)
			return AuthResponse{}, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		userId, err := s.Authenticate(ctx, entities.User{
			Name:     req.Name,
			Password: req.Pwd,
		})
		if err != nil {
			level.Error(logger).Log("error", err.Error())
			return AuthResponse{}, err
		}

		return AuthResponse{
			UserId: userId,
		}, nil
	}
}

func makeCreateUserEndpoint(s Service, logger log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		logger := log.With(logger, "method", "makeCreateUserEndpoint")
		req, ok := request.(CreateUserRequest)
		if !ok {
			level.Error(logger).Log("error", erro.ErrInvalidRequestType)
			return CreateUserResponse{}, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		res, err := s.CreateUser(ctx, entities.User{
			Name:     req.Name,
			Password: req.Pwd,
			Age:      req.Age,
			AddInfo:  sql.NullString{String: req.AddInfo},
		})
		if err != nil {
			level.Error(logger).Log("error", err.Error())
			return CreateUserResponse{}, err
		}

		return CreateUserResponse{
			UserId: res.UserId,
		}, err
	}
}

func makeUpdateUserEndpoint(s Service, logger log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		logger := log.With(logger, "method", "makeUpdateUserEndpoint")
		req, ok := request.(UpdateUserRequest)
		if !ok {
			level.Error(logger).Log("error", erro.ErrInvalidRequestType)
			return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		err := s.UpdateUser(ctx, entities.User{
			UserId:   req.UserId,
			Name:     req.Name,
			Password: req.Pwd,
			Age:      req.Age,
			AddInfo:  sql.NullString{String: req.AddInfo},
		})
		if err != nil {
			level.Error(logger).Log("error", err.Error())
			return nil, err
		}

		return nil, nil
	}
}

func makeGetUserEndpoint(s Service, logger log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		logger := log.With(logger, "method", "makeGetUserEndpoint")
		req, ok := request.(GetUserRequest)
		if !ok {
			level.Error(logger).Log("error", erro.ErrInvalidRequestType)
			return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		user, err := s.GetUser(ctx, req.UserId)
		if err != nil {
			level.Error(logger).Log("error", err.Error())
			return GetUserResponse{}, err
		}

		return GetUserResponse{
			UserId:  user.UserId,
			Name:    user.Name,
			Age:     user.Age,
			AddInfo: user.AddInfo.String,
		}, nil
	}
}
