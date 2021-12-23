package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	erro "github.com/javibauza/final-project/grpc-service/errors"
	"github.com/javibauza/final-project/grpc-service/service"
)

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

func MakeEndpoints(s service.Service) Endpoints {
	return Endpoints{
		Authenticate: makeAuthEndpoint(s),
		CreateUser:   makeCreateUserEndpoint(s),
		UpdateUser:   makeUpdateUserEndpoint(s),
		GetUser:      makeGetUserEndpoint(s),
	}
}

func makeAuthEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(AuthRequest)
		if !ok {
			return AuthResponse{}, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		userId, err := s.Authenticate(ctx, service.AuthRequest{
			Name: req.Name,
			Pwd:  req.Pwd,
		})
		if err != nil {
			return AuthResponse{}, err
		}

		return AuthResponse{
			UserId: userId,
		}, nil
	}
}

func makeCreateUserEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(CreateUserRequest)
		if !ok {
			return CreateUserResponse{}, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		res, err := s.CreateUser(ctx, service.CreateUserRequest{
			Name:    req.Name,
			Pwd:     req.Pwd,
			Age:     req.Age,
			AddInfo: req.AddInfo,
		})
		if err != nil {
			return CreateUserResponse{}, err
		}

		return CreateUserResponse{
			UserId: res.UserId,
		}, err
	}
}

func makeUpdateUserEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(UpdateUserRequest)
		if !ok {
			return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		err := s.UpdateUser(ctx, service.UpdateUserRequest{
			UserId:  req.UserId,
			Name:    req.Name,
			Pwd:     req.Pwd,
			Age:     req.Age,
			AddInfo: req.AddInfo,
		})
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func makeGetUserEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetUserRequest)
		if !ok {
			return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		user, err := s.GetUser(ctx, req.UserId)
		if err != nil {
			return GetUserResponse{}, err
		}

		return GetUserResponse{
			UserId:  user.UserId,
			Name:    user.Name,
			Age:     user.Age,
			AddInfo: user.AddInfo,
		}, nil
	}
}
