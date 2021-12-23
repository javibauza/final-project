package transport

import (
	"context"
	"fmt"

	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"

	"github.com/javibauza/final-project/grpc-service/endpoints"
	erro "github.com/javibauza/final-project/grpc-service/errors"
	"github.com/javibauza/final-project/grpc-service/pb"
)

type gRPCServer struct {
	auth       gt.Handler
	createUser gt.Handler
	updateUser gt.Handler
	getUser    gt.Handler
	pb.UnimplementedUserServiceServer
}

func NewGRPCServer(endpoints endpoints.Endpoints, logger log.Logger) pb.UserServiceServer {
	return &gRPCServer{
		auth: gt.NewServer(
			endpoints.Authenticate,
			decodeAuthRequest,
			encodeAuthResponse,
		),
		createUser: gt.NewServer(
			endpoints.CreateUser,
			decodeCreateUserRequest,
			encodeCreateUserResponse,
		),
		updateUser: gt.NewServer(
			endpoints.UpdateUser,
			decodeUpdateUserRequest,
			encodeUpdateUserResponse,
		),
		getUser: gt.NewServer(
			endpoints.GetUser,
			decodeGetUserRequest,
			encodeGetUserResponse,
		),
	}
}

func (s *gRPCServer) Authenticate(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	_, resp, err := s.auth.ServeGRPC(ctx, req)
	if err != nil {
		var status = &pb.Status{}
		var authResponse = &pb.AuthResponse{}

		switch r := err.(type) {
		case *erro.ErrInvalidArgument,
			*erro.ErrNotFound,
			*erro.ErrPermissionDenied:
			status = resolveStatus(r)
		default:
			status.Code = 3
			status.Message = fmt.Sprintf("unexpected error: %s", err.Error())
		}

		authResponse.Status = status
		return authResponse, nil
	}

	authResp, ok := resp.(*pb.AuthResponse)
	if !ok {
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}

	return authResp, nil
}

func decodeAuthRequest(_ context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(*pb.AuthRequest)
	if !ok {
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}
	return endpoints.AuthRequest{Pwd: req.Password, Name: req.UserName}, nil
}

func encodeAuthResponse(_ context.Context, response interface{}) (interface{}, error) {
	var status = &pb.Status{}
	var authResponse = &pb.AuthResponse{}
	switch r := response.(type) {
	case endpoints.AuthResponse:
		status.Code = 0
		status.Message = "ok"
		authResponse.UserId = r.UserId
	default:
		status.Code = 3
		status.Message = "unexpected error"
	}

	authResponse.Status = status
	return authResponse, nil
}

func (s *gRPCServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	_, res, err := s.createUser.ServeGRPC(ctx, req)
	if err != nil {
		var status = &pb.Status{}
		var createUserResponse = &pb.CreateUserResponse{}

		switch r := err.(type) {
		case *erro.ErrInvalidArgument,
			*erro.ErrNotFound,
			*erro.ErrPermissionDenied:
			status = resolveStatus(r)
		default:
			status.Code = 3
			status.Message = fmt.Sprintf("unexpected error: %s", err.Error())
		}

		createUserResponse.Status = status
		return createUserResponse, nil
	}

	createRes, ok := res.(*pb.CreateUserResponse)
	if !ok {
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}

	return createRes, nil
}

func decodeCreateUserRequest(_ context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(*pb.CreateUserRequest)
	if !ok {
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}
	return endpoints.CreateUserRequest{
		Name:    req.UserName,
		Pwd:     req.Password,
		Age:     req.UserAge,
		AddInfo: req.AddInfo,
	}, nil
}

func encodeCreateUserResponse(_ context.Context, response interface{}) (interface{}, error) {
	var status = &pb.Status{}
	var createUserResponse = &pb.CreateUserResponse{}
	switch r := response.(type) {
	case endpoints.CreateUserResponse:
		status.Code = 0
		status.Message = "ok"
		createUserResponse.UserId = r.UserId
	default:
		status.Code = 3
		status.Message = "unexpected error"
	}

	createUserResponse.Status = status
	return createUserResponse, nil
}

func (s *gRPCServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	_, res, err := s.updateUser.ServeGRPC(ctx, req)
	if err != nil {
		var status = &pb.Status{}
		var updateUserResponse = &pb.UpdateUserResponse{}

		switch r := err.(type) {
		case *erro.ErrInvalidArgument,
			*erro.ErrNotFound,
			*erro.ErrPermissionDenied:
			status = resolveStatus(r)
		default:
			status.Code = 3
			status.Message = fmt.Sprintf("unexpected error: %s", err.Error())
		}

		updateUserResponse.Status = status
		return updateUserResponse, nil
	}

	updateRes, ok := res.(*pb.UpdateUserResponse)
	if !ok {
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}

	return updateRes, nil
}

func decodeUpdateUserRequest(_ context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(*pb.UpdateUserRequest)
	if !ok {
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}
	return endpoints.UpdateUserRequest{
		UserId:  req.UserId,
		Name:    req.UserName,
		Pwd:     req.Password,
		Age:     req.UserAge,
		AddInfo: req.AddInfo,
	}, nil
}

func encodeUpdateUserResponse(_ context.Context, response interface{}) (interface{}, error) {
	var status = &pb.Status{}
	var updateUserResponse = &pb.UpdateUserResponse{}
	switch response.(type) {
	case nil:
		status.Code = 0
		status.Message = "ok"
	default:
		status.Code = 3
		status.Message = "unexpected error"
	}

	updateUserResponse.Status = status
	return updateUserResponse, nil
}

func (s *gRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	_, res, err := s.getUser.ServeGRPC(ctx, req)
	if err != nil {
		var status = &pb.Status{}
		var getUserResponse = &pb.GetUserResponse{}

		switch r := err.(type) {
		case *erro.ErrInvalidArgument,
			*erro.ErrNotFound,
			*erro.ErrPermissionDenied:
			status = resolveStatus(r)
		default:
			status.Code = 3
			status.Message = fmt.Sprintf("unexpected error: %s", err.Error())
		}

		getUserResponse.Status = status
		return getUserResponse, nil
	}

	response, ok := res.(*pb.GetUserResponse)
	if !ok {
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}

	return response, nil
}

func decodeGetUserRequest(_ context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(*pb.GetUserRequest)
	if !ok {
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}
	return endpoints.GetUserRequest{
		UserId: req.UserId,
	}, nil
}

func encodeGetUserResponse(_ context.Context, response interface{}) (interface{}, error) {
	var status = &pb.Status{}
	var getUserResponse = &pb.GetUserResponse{}
	switch r := response.(type) {
	case endpoints.GetUserResponse:
		status.Code = 0
		status.Message = "ok"
		getUserResponse.UserId = r.UserId
		getUserResponse.UserName = r.Name
		getUserResponse.UserAge = r.Age
		getUserResponse.AddInfo = r.AddInfo
	default:
		status.Code = 3
		status.Message = "unexpected error"
	}

	getUserResponse.Status = status
	return getUserResponse, nil
}

func resolveStatus(response interface{}) *pb.Status {
	var status pb.Status
	switch r := response.(type) {
	case *erro.ErrInvalidArgument:
		status.Code = 3
		status.Message = r.Err.Error()
	case *erro.ErrNotFound:
		status.Code = 5
		status.Message = r.Err.Error()
	case *erro.ErrPermissionDenied:
		status.Code = 7
		status.Message = r.Err.Error()
	default:
		status.Code = 3
		status.Message = "unexpected error"
	}

	return &status
}
