package repository

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	gokitLog "github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/javibauza/final-project/grpc-service/pb"
	"github.com/javibauza/grpc-service/utils"
)

type mockGRPCService struct {
	mock.Mock
	pb.UnimplementedUserServiceServer
}

func (m *mockGRPCService) Authenticate(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*pb.AuthResponse), args.Error(1)
}

func (m *mockGRPCService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*pb.CreateUserResponse), args.Error(1)
}

func (m *mockGRPCService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*pb.UpdateUserResponse), args.Error(1)
}

func (m *mockGRPCService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*pb.GetUserResponse), args.Error(1)
}

func dialer(m *mockGRPCService) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterUserServiceServer(server, m)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestAuthenticate(t *testing.T) {
	var logger gokitLog.Logger
	{
		logger = gokitLog.NewLogfmtLogger(os.Stderr)
		logger = gokitLog.NewSyncLogger(logger)
		logger = gokitLog.With(logger,
			"service", "repository_test",
			"time:", gokitLog.DefaultTimestampUTC,
			"caller", gokitLog.DefaultCaller,
		)
	}

	ctx := context.Background()

	grpcUserService := new(mockGRPCService)
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(grpcUserService)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	userRepoSvc := NewUserRepo(conn, logger)

	testCases := []struct {
		testName      string
		userId        string
		request       User
		grpcRequest   func(req User) *pb.AuthRequest
		grpcResponse  func(userId string) (*pb.AuthResponse, error)
		checkResponse func(t *testing.T, userId string, response User, resError error)
	}{
		{
			testName: "user authenticated",
			request: User{
				Name:     "javier",
				Password: "javier123",
			},
			userId: utils.RandomString(12),
			grpcRequest: func(req User) *pb.AuthRequest {
				return &pb.AuthRequest{
					UserName: req.Name,
					Password: req.Password,
				}
			},
			grpcResponse: func(userId string) (*pb.AuthResponse, error) {
				status := &pb.Status{
					Code:    0,
					Message: "ok",
				}
				return &pb.AuthResponse{
					UserId: userId,
					Status: status,
				}, nil
			},
			checkResponse: func(t *testing.T, userId string, response User, resError error) {
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user not found",
			request: User{
				Name:     "reivaj",
				Password: "javier123",
			},
			userId: utils.RandomString(12),
			grpcRequest: func(req User) *pb.AuthRequest {
				return &pb.AuthRequest{
					UserName: req.Name,
					Password: req.Password,
				}
			},
			grpcResponse: func(userId string) (*pb.AuthResponse, error) {
				status := &pb.Status{
					Code:    5,
					Message: "user not found",
				}
				return &pb.AuthResponse{
					Status: status,
				}, nil
			},
			checkResponse: func(t *testing.T, userId string, response User, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, "user not found")
			},
		},
		{
			testName: "wrong password",
			request: User{
				Name:     "javier",
				Password: "javier321",
			},
			userId: utils.RandomString(12),
			grpcRequest: func(req User) *pb.AuthRequest {
				return &pb.AuthRequest{
					UserName: req.Name,
					Password: req.Password,
				}
			},
			grpcResponse: func(userId string) (*pb.AuthResponse, error) {
				status := &pb.Status{
					Code:    7,
					Message: "wrong password",
				}
				return &pb.AuthResponse{
					Status: status,
				}, nil
			},
			checkResponse: func(t *testing.T, userId string, response User, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, "wrong password")
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			grpcResponse, err := tc.grpcResponse(tc.userId)
			grpcUserService.On("Authenticate", mock.Anything, tc.grpcRequest(tc.request)).
				Return(grpcResponse, err)

			res, err := userRepoSvc.Authenticate(ctx, tc.request)
			tc.checkResponse(t, tc.userId, res, err)
		})
	}
}

func TestCreateUser(t *testing.T) {
	var logger gokitLog.Logger
	{
		logger = gokitLog.NewLogfmtLogger(os.Stderr)
		logger = gokitLog.NewSyncLogger(logger)
		logger = gokitLog.With(logger,
			"service", "service_test",
			"time:", gokitLog.DefaultTimestampUTC,
			"caller", gokitLog.DefaultCaller,
		)
	}

	ctx := context.Background()

	grpcUserService := new(mockGRPCService)
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(grpcUserService)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	userRepoSvc := NewUserRepo(conn, logger)

	testCases := []struct {
		testName      string
		userId        string
		request       User
		grpcRequest   func(req User) *pb.CreateUserRequest
		grpcResponse  func(userId string) (*pb.CreateUserResponse, error)
		checkResponse func(t *testing.T, userId, response string, resError error)
	}{
		{
			testName: "user created",
			request: User{
				Name:     "javier",
				Password: "javier123",
				Age:      45,
				AddInfo:  "",
			},
			userId: utils.RandomString(12),
			grpcRequest: func(req User) *pb.CreateUserRequest {
				return &pb.CreateUserRequest{
					UserName: req.Name,
					Password: req.Password,
					UserAge:  req.Age,
					AddInfo:  req.AddInfo,
				}
			},
			grpcResponse: func(userId string) (*pb.CreateUserResponse, error) {
				status := &pb.Status{
					Code:    0,
					Message: "ok",
				}
				return &pb.CreateUserResponse{
					UserId: userId,
					Status: status,
				}, nil
			},
			checkResponse: func(t *testing.T, userId, response string, resError error) {
				assert.Equal(t, userId, response)
				assert.NoError(t, resError)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			if tc.grpcRequest != nil {
				grpcResponse, err := tc.grpcResponse(tc.userId)
				grpcUserService.On("CreateUser", mock.Anything, tc.grpcRequest(tc.request)).
					Return(grpcResponse, err)
			}

			res, err := userRepoSvc.CreateUser(ctx, tc.request)
			tc.checkResponse(t, tc.userId, res, err)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	var logger gokitLog.Logger
	{
		logger = gokitLog.NewLogfmtLogger(os.Stderr)
		logger = gokitLog.NewSyncLogger(logger)
		logger = gokitLog.With(logger,
			"service", "service_test",
			"time:", gokitLog.DefaultTimestampUTC,
			"caller", gokitLog.DefaultCaller,
		)
	}

	ctx := context.Background()

	grpcUserService := new(mockGRPCService)
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(grpcUserService)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	userRepoSvc := NewUserRepo(conn, logger)

	testCases := []struct {
		testName      string
		request       User
		grpcRequest   func(req User) *pb.UpdateUserRequest
		grpcResponse  func() (*pb.UpdateUserResponse, error)
		checkResponse func(t *testing.T, resError error)
	}{
		{
			testName: "user updated",
			request: User{
				Name:     "javier",
				Password: "javier123",
				Age:      45,
				AddInfo:  "",
			},
			grpcRequest: func(req User) *pb.UpdateUserRequest {
				return &pb.UpdateUserRequest{
					UserName: req.Name,
					Password: req.Password,
					UserAge:  req.Age,
					AddInfo:  req.AddInfo,
				}
			},
			grpcResponse: func() (*pb.UpdateUserResponse, error) {
				status := &pb.Status{
					Code:    0,
					Message: "ok",
				}
				return &pb.UpdateUserResponse{
					Status: status,
				}, nil
			},
			checkResponse: func(t *testing.T, resError error) {
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user not found",
			request: User{
				Name:     "reivaj",
				Password: "javier123",
				Age:      37,
				AddInfo:  "Some test additional info",
			},
			grpcRequest: func(req User) *pb.UpdateUserRequest {
				return &pb.UpdateUserRequest{
					UserName: req.Name,
					Password: req.Password,
					UserAge:  req.Age,
					AddInfo:  req.AddInfo,
				}
			},
			grpcResponse: func() (*pb.UpdateUserResponse, error) {
				status := &pb.Status{
					Code:    5,
					Message: "user not found",
				}
				return &pb.UpdateUserResponse{
					Status: status,
				}, nil
			},
			checkResponse: func(t *testing.T, resError error) {
				assert.EqualError(t, resError, "user not found")
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			if tc.grpcRequest != nil {
				res, err := tc.grpcResponse()
				grpcUserService.On("UpdateUser", mock.Anything, tc.grpcRequest(tc.request)).
					Return(res, err)
			}

			err := userRepoSvc.UpdateUser(ctx, tc.request)
			tc.checkResponse(t, err)
		})
	}
}

func TestGetUser(t *testing.T) {
	var logger gokitLog.Logger
	{
		logger = gokitLog.NewLogfmtLogger(os.Stderr)
		logger = gokitLog.NewSyncLogger(logger)
		logger = gokitLog.With(logger,
			"service", "service_test",
			"time:", gokitLog.DefaultTimestampUTC,
			"caller", gokitLog.DefaultCaller,
		)
	}

	ctx := context.Background()

	grpcUserService := new(mockGRPCService)
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer(grpcUserService)))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	userRepoSvc := NewUserRepo(conn, logger)

	testCases := []struct {
		testName      string
		userId        string
		grpcRequest   func(req User) *pb.GetUserRequest
		grpcResponse  func(user User) (*pb.GetUserResponse, error)
		checkResponse func(t *testing.T, res User, resError error)
	}{
		{
			testName: "user obtained",
			userId:   utils.RandomString(12),
			grpcRequest: func(req User) *pb.GetUserRequest {
				return &pb.GetUserRequest{
					UserId: req.UserId,
				}
			},
			grpcResponse: func(user User) (*pb.GetUserResponse, error) {
				status := &pb.Status{
					Code:    0,
					Message: "ok",
				}
				return &pb.GetUserResponse{
					Status:   status,
					UserId:   user.UserId,
					UserName: user.Name,
					UserAge:  user.Age,
					AddInfo:  user.AddInfo,
				}, nil
			},
			checkResponse: func(t *testing.T, res User, resError error) {
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user not found",
			userId:   utils.RandomString(12),
			grpcRequest: func(req User) *pb.GetUserRequest {
				return &pb.GetUserRequest{
					UserId: req.UserId,
				}
			},
			grpcResponse: func(user User) (*pb.GetUserResponse, error) {
				status := &pb.Status{
					Code:    5,
					Message: "user not found",
				}
				return &pb.GetUserResponse{
					Status: status,
				}, nil
			},
			checkResponse: func(t *testing.T, res User, resError error) {
				assert.EqualError(t, resError, "user not found")
			},
		},
	}

	user := User{
		Name:    "javier",
		Age:     37,
		AddInfo: "Some test additional info",
	}

	for i := range testCases {
		tc := testCases[i]
		user.UserId = tc.userId
		t.Run(tc.testName, func(t *testing.T) {
			res, err := tc.grpcResponse(user)
			grpcUserService.On("GetUser", mock.Anything, tc.grpcRequest(user)).
				Return(res, err)
			repoResponse, err := userRepoSvc.GetUser(ctx, tc.userId)
			tc.checkResponse(t, repoResponse, err)
		})
	}
}
