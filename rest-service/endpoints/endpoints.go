package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/javibauza/final-project/rest-service/service"

	erro "github.com/javibauza/final-project/rest-service/errors"
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
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(AuthRequest)
		if !ok {
			return nil, erro.NewErrBadRequest(erro.ErrInvalidInputType)
		}

		res, err := s.Authenticate(ctx, service.AuthRequest{Name: req.Name, Pwd: req.Pwd})
		if err != nil {
			return AuthResponse{}, err
		}

		return AuthResponse{
			UserId: res.UserId,
		}, nil
	}
}

func makeCreateUserEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(CreateUserRequest)
		if !ok {
			return nil, erro.NewErrBadRequest(erro.ErrInvalidInputType)
		}

		res, err := s.CreateUser(ctx, service.CreateUserRequest{
			Name:    req.Name,
			Pwd:     req.Pwd,
			Age:     req.Age,
			AddInfo: req.AddInfo,
		})
		if err != nil {
			return nil, err
		}

		return CreateUserResponse{
			UserId: res.UserId,
		}, nil
	}
}

func makeUpdateUserEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(UpdateUserRequest)
		if !ok {
			return nil, erro.NewErrBadRequest(erro.ErrInvalidInputType)
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
			return nil, erro.NewErrBadRequest(erro.ErrInvalidInputType)
		}
		user, err := s.GetUser(ctx, req.UserId)
		if err != nil {
			return nil, err
		}

		return GetUserResponse{
			UserId:  user.UserId,
			Name:    user.Name,
			Age:     user.Age,
			AddInfo: user.AddInfo,
		}, nil
	}
}
