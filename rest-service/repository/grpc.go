package repository

import (
	"context"
	"errors"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"google.golang.org/grpc"

	"../github.com/javibauza/final-project/grpc-service/pb"

	erro "github.com/javibauza/final-project/gbfp-httpservice/errors"
)

type UserRepo struct {
	conn   *grpc.ClientConn
	logger log.Logger
}

type UserRepository interface {
	Authenticate(ctx context.Context, user User) (User, error)
	CreateUser(ctx context.Context, user User) (string, error)
	UpdateUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, userId string) (User, error)
}

type User struct {
	UserId   string
	Name     string
	Password string
	Age      uint32
	AddInfo  string
}

func NewUserRepo(conn *grpc.ClientConn, logger log.Logger) UserRepository {
	return &UserRepo{
		conn:   conn,
		logger: log.With(logger, "error", "grpc"),
	}
}

func (r *UserRepo) Authenticate(ctx context.Context, user User) (User, error) {
	logger := log.With(r.logger, "method", "Authenticate")

	request := &pb.AuthRequest{
		UserName: user.Name,
		Password: user.Password,
	}

	client := pb.NewUserServiceClient(r.conn)
	grpcResponse, err := client.Authenticate(ctx, request)

	if err != nil {
		level.Error(logger).Log("err", err)
		return User{}, err
	}

	if grpcResponse.Status.Code == 0 {
		return User{UserId: grpcResponse.UserId}, nil
	} else {
		code := grpcResponse.Status.Code
		message := grpcResponse.Status.Message
		level.Info(logger).Log("grpc response code", code, "grpc response message", message)
		return User{}, grpcErrorHandler(code, message)
	}
}

func (r *UserRepo) CreateUser(ctx context.Context, user User) (string, error) {
	logger := log.With(r.logger, "method", "CreateUser")

	request := pb.CreateUserRequest{
		UserName: user.Name,
		Password: user.Password,
		UserAge:  user.Age,
		AddInfo:  user.AddInfo,
	}

	client := pb.NewUserServiceClient(r.conn)
	grpcResponse, err := client.CreateUser(ctx, &request)
	if err != nil {
		level.Error(logger).Log("err", err)
		return "", err
	}

	if grpcResponse.Status.Code == 0 {
		return grpcResponse.UserId, nil
	} else {
		return "", grpcErrorHandler(grpcResponse.Status.Code, grpcResponse.Status.Message)
	}
}

func (r *UserRepo) UpdateUser(ctx context.Context, user User) error {
	logger := log.With(r.logger, "method", "UpdateUser")

	request := pb.UpdateUserRequest{
		UserId:   user.UserId,
		UserName: user.Name,
		Password: user.Password,
		UserAge:  user.Age,
		AddInfo:  user.AddInfo,
	}

	client := pb.NewUserServiceClient(r.conn)
	grpcResponse, err := client.UpdateUser(ctx, &request)
	if err != nil {
		level.Error(logger).Log("err", err)
		return err
	}

	if grpcResponse.Status.Code == 0 {
		return nil
	} else {
		return grpcErrorHandler(grpcResponse.Status.Code, grpcResponse.Status.Message)
	}
}

func (r *UserRepo) GetUser(ctx context.Context, userId string) (User, error) {
	logger := log.With(r.logger, "method", "UpdateUser")

	request := pb.GetUserRequest{
		UserId: userId,
	}

	client := pb.NewUserServiceClient(r.conn)
	grpcResponse, err := client.GetUser(ctx, &request)
	if err != nil {
		level.Error(logger).Log("err", err)
		return User{}, err
	}

	resCode := grpcResponse.Status.Code
	resMessage := grpcResponse.Status.Message
	if resCode == 0 {
		return User{
			UserId:  grpcResponse.UserId,
			Name:    grpcResponse.UserName,
			Age:     grpcResponse.UserAge,
			AddInfo: grpcResponse.AddInfo,
		}, nil
	} else {
		level.Error(logger).Log("grpc status code", resCode, "grpc status message", resMessage)
		return User{}, grpcErrorHandler(resCode, resMessage)
	}
}

func grpcErrorHandler(code int32, message string) error {
	err := errors.New(message)
	switch code {
	case 2:
		return erro.ErrInternal{Err: err}
	case 3:
		return erro.ErrBadRequest{Err: err}
	case 5:
		return erro.ErrNotFound{Err: err}
	case 7:
		return erro.ErrForbidden{Err: err}
	default:
		return erro.ErrInternal{Err: errors.New(erro.ErrUnexpected)}
	}
}
