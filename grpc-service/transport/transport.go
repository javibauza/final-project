package transport

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/javibauza/final-project/grpc-service/endpoints"
	erro "github.com/javibauza/final-project/grpc-service/errors"
	pb "github.com/javibauza/final-project/grpc-service/proto"
)

type gRPCServer struct {
	logger     log.Logger
	auth       gt.Handler
	createUser gt.Handler
	updateUser gt.Handler
	getUser    gt.Handler
	pb.UnimplementedUserServiceServer
}

func NewGRPCServer(endpoints endpoints.Endpoints, logger log.Logger) pb.UserServiceServer {
	return &gRPCServer{
		logger: logger,
		auth: gt.NewServer(
			endpoints.Authenticate,
			makeDecodeAuthenticateRequest(logger),
			makeEncodeAuthenticateResponse(logger),
		),
		createUser: gt.NewServer(
			endpoints.CreateUser,
			makeDecodeCreateUserRequest(logger),
			makeEncodeCreateUserResponse(logger),
		),
		updateUser: gt.NewServer(
			endpoints.UpdateUser,
			makeDecodeUpdateUserRequest(logger),
			makeEncodeUpdateUserResponse(logger),
		),
		getUser: gt.NewServer(
			endpoints.GetUser,
			makeDecodeGetUserRequest(logger),
			makeEncodeGetUserResponse(logger),
		),
	}
}

func (s *gRPCServer) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	logger := log.With(s.logger, "method", "Authenticate")
	_, resp, err := s.auth.ServeGRPC(ctx, req)
	if err != nil {
		var AuthenticateResponse = &pb.AuthenticateResponse{}
		status := setStatus(ctx, err)

		level.Info(logger).Log("response", err.Error())
		AuthenticateResponse.Status = status
		return AuthenticateResponse, nil
	}

	authResp, ok := resp.(*pb.AuthenticateResponse)
	if !ok {
		level.Error(logger).Log("error", erro.ErrInvalidRequestType)
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}

	return authResp, nil
}

func makeDecodeAuthenticateRequest(logger log.Logger) func(_ context.Context, request interface{}) (interface{}, error) {
	logger = log.With(logger, "method", "makeDecodeAuthenticateRequest")

	return func(_ context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*pb.AuthenticateRequest)
		if !ok {
			level.Error(logger).Log("error", erro.ErrInvalidRequestType)
			return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}
		return endpoints.AuthRequest{Pwd: req.Password, Name: req.UserName}, nil
	}
}

func makeEncodeAuthenticateResponse(logger log.Logger) func(ctx context.Context, response interface{}) (interface{}, error) {
	logger = log.With(logger, "method", "makeEncodeAuthenticateRequest")

	return func(ctx context.Context, response interface{}) (interface{}, error) {
		var status = &pb.Status{}
		var AuthenticateResponse = &pb.AuthenticateResponse{}
		switch r := response.(type) {
		case endpoints.AuthResponse:
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-codes", "200"))
			status.Code = 0
			status.Message = "ok"
			AuthenticateResponse.UserId = r.UserId
			level.Info(logger).Log("response", "response ok")
		default:
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-codes", "500"))
			status.Code = 3
			status.Message = "unexpected error"
			level.Info(logger).Log("response", "unexpected error")
		}

		AuthenticateResponse.Status = status
		return AuthenticateResponse, nil
	}
}

func (s *gRPCServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	logger := log.With(s.logger, "method", "CreateUser")

	_, res, err := s.createUser.ServeGRPC(ctx, req)
	if err != nil {
		var createUserResponse = &pb.CreateUserResponse{}
		status := setStatus(ctx, err)

		level.Info(logger).Log("response", err.Error())
		createUserResponse.Status = status
		return createUserResponse, nil
	}

	createRes, ok := res.(*pb.CreateUserResponse)
	if !ok {
		level.Error(logger).Log("error", erro.ErrInvalidRequestType)
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}

	return createRes, nil
}

func makeDecodeCreateUserRequest(logger log.Logger) func(_ context.Context, request interface{}) (interface{}, error) {
	log.With(logger, "method", "makeDecodeCreateUserRequest")

	return func(_ context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*pb.CreateUserRequest)
		if !ok {
			level.Error(logger).Log("error", erro.ErrInvalidRequestType)
			return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}

		return endpoints.CreateUserRequest{
			Name:    req.UserName,
			Pwd:     req.Password,
			Age:     req.UserAge,
			AddInfo: req.AddInfo,
		}, nil
	}
}

func makeEncodeCreateUserResponse(logger log.Logger) func(_ context.Context, response interface{}) (interface{}, error) {
	log.With(logger, "method", "makeEncodeCreateUserResponse")

	return func(ctx context.Context, response interface{}) (interface{}, error) {
		var status = &pb.Status{}
		var createUserResponse = &pb.CreateUserResponse{}
		switch r := response.(type) {
		case endpoints.CreateUserResponse:
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-codes", "200"))
			status.Code = 0
			status.Message = "ok"
			createUserResponse.UserId = r.UserId
		default:
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-codes", "500"))
			status.Code = 3
			status.Message = "unexpected error"
		}

		createUserResponse.Status = status
		return createUserResponse, nil
	}
}

func (s *gRPCServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	logger := log.With(s.logger, "method", "UpdateUser")

	_, res, err := s.updateUser.ServeGRPC(ctx, req)
	if err != nil {
		var updateUserResponse = &pb.UpdateUserResponse{}
		status := setStatus(ctx, err)

		level.Info(logger).Log("response", err.Error())
		updateUserResponse.Status = status
		return updateUserResponse, nil
	}

	updateRes, ok := res.(*pb.UpdateUserResponse)
	if !ok {
		level.Error(logger).Log("error", erro.ErrInvalidRequestType)
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}

	return updateRes, nil
}

func makeDecodeUpdateUserRequest(logger log.Logger) func(_ context.Context, request interface{}) (interface{}, error) {
	log.With(logger, "method", "makeDecodeUpdateUserRequest")

	return func(_ context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*pb.UpdateUserRequest)
		if !ok {
			level.Error(logger).Log("error", erro.ErrInvalidRequestType)
			return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}
		level.Debug(logger).Log("request", fmt.Sprintf("%v", request))
		return endpoints.UpdateUserRequest{
			UserId:  req.User.UserId,
			Name:    req.User.UserName,
			Pwd:     req.User.Password,
			Age:     req.User.UserAge,
			AddInfo: req.User.AddInfo,
		}, nil
	}
}

func makeEncodeUpdateUserResponse(logger log.Logger) func(_ context.Context, response interface{}) (interface{}, error) {
	log.With(logger, "method", "makeEncodeUpdateUserResponse")

	return func(ctx context.Context, response interface{}) (interface{}, error) {
		var status = &pb.Status{}
		var updateUserResponse = &pb.UpdateUserResponse{}
		switch response.(type) {
		case nil:
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-codes", "200"))
			status.Code = 0
			status.Message = "ok"
		default:
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-codes", "500"))
			status.Code = 3
			status.Message = "unexpected error"
		}

		updateUserResponse.Status = status
		return updateUserResponse, nil
	}
}

func (s *gRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	logger := log.With(s.logger, "method", "GetUser")

	_, res, err := s.getUser.ServeGRPC(ctx, req)
	if err != nil {
		var getUserResponse = &pb.GetUserResponse{}
		status := setStatus(ctx, err)

		getUserResponse.Status = status
		return getUserResponse, nil
	}

	response, ok := res.(*pb.GetUserResponse)
	if !ok {
		level.Error(logger).Log("error", erro.ErrInvalidRequestType)
		return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
	}

	return response, nil
}

func makeDecodeGetUserRequest(logger log.Logger) func(_ context.Context, request interface{}) (interface{}, error) {
	log.With(logger, "method", "makeDecodeGetUserRequest")

	return func(_ context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*pb.GetUserRequest)
		if !ok {
			level.Error(logger).Log("error", erro.ErrInvalidRequestType)
			return nil, erro.NewErrInvalidArgument(erro.ErrInvalidRequestType)
		}
		return endpoints.GetUserRequest{
			UserId: req.UserId,
		}, nil
	}
}

func makeEncodeGetUserResponse(logger log.Logger) func(_ context.Context, response interface{}) (interface{}, error) {
	log.With(logger, "method", "makeEncodeGetUserResponse")

	return func(ctx context.Context, response interface{}) (interface{}, error) {
		var status = &pb.Status{}
		var getUserResponse = &pb.GetUserResponse{}
		switch r := response.(type) {
		case endpoints.GetUserResponse:
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-codes", "200"))
			status.Code = 0
			status.Message = "ok"
			getUserResponse.UserId = r.UserId
			getUserResponse.UserName = r.Name
			getUserResponse.UserAge = r.Age
			getUserResponse.AddInfo = r.AddInfo
		default:
			_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-codes", "500"))
			status.Code = 3
			status.Message = "unexpected error"
		}

		getUserResponse.Status = status
		return getUserResponse, nil
	}
}

func setStatus(ctx context.Context, e error) *pb.Status {
	var status = &pb.Status{}

	switch r := e.(type) {
	case *erro.ErrInvalidArgument:
		grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "400"))
		status.Code = 3
		status.Message = r.Err.Error()
	case *erro.ErrNotFound:
		grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "404"))
		status.Code = 5
		status.Message = r.Err.Error()
	case *erro.ErrPermissionDenied:
		grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "403"))
		status.Code = 7
		status.Message = r.Err.Error()
	default:
		grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "500"))
		status.Code = 3
		status.Message = fmt.Sprintf("unexpected error: %s", r.Error())
	}

	return status
}

func MakeHTTPResponseModifier(logger log.Logger) func(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
	return func(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
		md, ok := runtime.ServerMetadataFromContext(ctx)
		if !ok {
			return nil
		}

		vals := md.HeaderMD.Get("x-http-code")

		if len(vals) > 0 {
			code, err := strconv.Atoi(vals[0])
			if err != nil {
				return err
			}

			delete(md.HeaderMD, "x-http-code")
			delete(w.Header(), "Grpc-Metadata-X-Http-Code")
			w.WriteHeader(code)
		}

		return nil
	}
}
