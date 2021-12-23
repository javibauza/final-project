package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/go-kit/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	erro "github.com/javibauza/final-project/gbfp-httpservice/errors"
	"github.com/javibauza/final-project/gbfp-httpservice/repository"
	"github.com/javibauza/final-project/gbfp-httpservice/utils"
)

type repoMock struct {
	mock.Mock
}

func (m *repoMock) Authenticate(ctx context.Context, req repository.User) (repository.User, error) {
	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return repository.User{}, args.Error(1)
	}

	return args.Get(0).(repository.User), args.Error(1)
}

func (m *repoMock) CreateUser(ctx context.Context, user repository.User) (string, error) {
	args := m.Called(ctx, user)

	return args.String(0), args.Error(1)
}

func (m *repoMock) UpdateUser(ctx context.Context, user repository.User) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *repoMock) GetUser(ctx context.Context, userId string) (repository.User, error) {
	args := m.Called(ctx, userId)

	return args.Get(0).(repository.User), args.Error(1)
}

func TestAuthenticate(t *testing.T) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "service_test",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	repoSvc := new(repoMock)

	service := NewService(repoSvc, logger)

	testCases := []struct {
		testName      string
		userName      string
		userPwd       string
		userId        string
		userPwdHash   []byte
		request       func(name, pwd string) AuthRequest
		repoResponse  func(userId string, pwdHash []byte) (repository.User, error)
		checkResponse func(t *testing.T, userId string, response AuthResponse, resError error)
	}{
		{
			testName:    "user authenticated",
			userName:    "javier",
			userPwd:     "javier123",
			userId:      utils.RandomString(12),
			userPwdHash: []byte("$2a$12$RXSLrffQZDUGljSPdQAPI.W4txkPKkeASl0qSM/tbx7mgMMqnDhui"),
			request: func(name, pwd string) AuthRequest {
				return AuthRequest{Name: name, Pwd: pwd}
			},
			repoResponse: func(userId string, pwdHash []byte) (repository.User, error) {
				return repository.User{UserId: userId}, nil
			},
			checkResponse: func(t *testing.T, userId string, response AuthResponse, resError error) {
				assert.Equal(t, response.UserId, userId)
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user not found",
			userName: "reivaj",
			userPwd:  "javier123",
			request: func(name, pwd string) AuthRequest {
				return AuthRequest{Name: name, Pwd: pwd}
			},
			repoResponse: func(userId string, pwdHash []byte) (repository.User, error) {
				return repository.User{}, &erro.ErrNotFound{Err: errors.New("user not found")}
			},
			checkResponse: func(t *testing.T, userId string, response AuthResponse, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, "user not found")
			},
		},
		{
			testName:    "wrong password",
			userName:    "javier",
			userPwd:     "javier321",
			userId:      utils.RandomString(12),
			userPwdHash: []byte("$2a$12$RXSLrffQZDUGljSPdQAPI.W4txkPKkeASl0qSM/tbx7mgMMqnDhui"),
			request: func(name, pwd string) AuthRequest {
				return AuthRequest{Name: name, Pwd: pwd}
			},
			repoResponse: func(userId string, pwdHash []byte) (repository.User, error) {
				return repository.User{}, &erro.ErrForbidden{Err: errors.New("wrong password")}
			},
			checkResponse: func(t *testing.T, userId string, response AuthResponse, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, "wrong password")
			},
		},
		{
			testName: "user name empty",
			userName: "",
			userPwd:  "javier321",
			request: func(name, pwd string) AuthRequest {
				return AuthRequest{Name: name, Pwd: pwd}
			},
			repoResponse: nil,
			checkResponse: func(t *testing.T, userId string, response AuthResponse, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, erro.ErrRequiredFields("name", "password"))
			},
		},
		{
			testName: "password empty",
			userName: "javier",
			userPwd:  "",
			request: func(name, pwd string) AuthRequest {
				return AuthRequest{Name: name, Pwd: pwd}
			},
			repoResponse: nil,
			checkResponse: func(t *testing.T, userId string, response AuthResponse, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, erro.ErrRequiredFields("name", "password"))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.repoResponse != nil {
				repoResponse, err := tc.repoResponse(tc.userId, tc.userPwdHash)
				repoSvc.On("Authenticate", ctx, repository.User{Name: tc.userName, Password: tc.userPwd}).
					Return(repoResponse, err)
			}

			res, err := service.Authenticate(ctx, tc.request(tc.userName, tc.userPwd))
			tc.checkResponse(t, tc.userId, res, err)
		})
	}
}

func TestCreateUser(t *testing.T) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "service_test",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	repoSvc := new(repoMock)

	service := NewService(repoSvc, logger)

	testCases := []struct {
		testName string
		userData struct {
			Name    string
			Pwd     string
			Age     uint32
			AddInfo string
		}
		userId        string
		request       func(name, pwd, addInfo string, age uint32) CreateUserRequest
		repoResponse  func(userId string) (string, error)
		checkResponse func(t *testing.T, userId string, response CreateUserResponse, resError error)
	}{
		{
			testName: "user created",
			userData: struct {
				Name    string
				Pwd     string
				Age     uint32
				AddInfo string
			}{
				"javier", "javier123", 45, ""},
			userId: utils.RandomString(12),
			request: func(name, pwd, addInfo string, age uint32) CreateUserRequest {
				return CreateUserRequest{
					Name:    name,
					Pwd:     pwd,
					Age:     age,
					AddInfo: addInfo,
				}
			},
			repoResponse: func(userId string) (string, error) {
				return userId, nil
			},
			checkResponse: func(t *testing.T, userId string, response CreateUserResponse, resError error) {
				assert.Equal(t, response.UserId, userId)
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user name empty",
			userData: struct {
				Name    string
				Pwd     string
				Age     uint32
				AddInfo string
			}{
				"", "javier123", 45, ""},
			userId: utils.RandomString(12),
			request: func(name, pwd, addInfo string, age uint32) CreateUserRequest {
				return CreateUserRequest{
					Name:    name,
					Pwd:     pwd,
					Age:     age,
					AddInfo: addInfo,
				}
			},
			repoResponse: func(userId string) (string, error) {
				return "", nil
			},
			checkResponse: func(t *testing.T, userId string, response CreateUserResponse, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, erro.ErrRequiredFields("name", "password"))
			},
		},
		{
			testName: "user password empty",
			userData: struct {
				Name    string
				Pwd     string
				Age     uint32
				AddInfo string
			}{
				"javier", "", 45, ""},
			userId: utils.RandomString(12),
			request: func(name, pwd, addInfo string, age uint32) CreateUserRequest {
				return CreateUserRequest{
					Name:    name,
					Pwd:     pwd,
					Age:     age,
					AddInfo: addInfo,
				}
			},
			repoResponse: func(userId string) (string, error) {
				return "", nil
			},
			checkResponse: func(t *testing.T, userId string, response CreateUserResponse, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, erro.ErrRequiredFields("name", "password"))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			userId, err := tc.repoResponse(tc.userId)
			repoSvc.On("CreateUser", ctx, repository.User{
				Name:     tc.userData.Name,
				Password: tc.userData.Pwd,
				Age:      tc.userData.Age,
				AddInfo:  tc.userData.AddInfo,
			}).
				Return(userId, err)
			res, err := service.CreateUser(ctx, tc.request(tc.userData.Name, tc.userData.Pwd, tc.userData.AddInfo, tc.userData.Age))
			tc.checkResponse(t, tc.userId, res, err)
		})
	}
}

func TestGetUser(t *testing.T) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "service_test",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	repoSvc := new(repoMock)

	service := NewService(repoSvc, logger)

	user := repository.User{
		UserId:  utils.RandomString(12),
		Name:    "javier",
		Age:     37,
		AddInfo: "some additional info",
	}

	testCases := []struct {
		testName string
		//userId        string
		repoResponse  func(userId string) (repository.User, error)
		checkResponse func(t *testing.T, res GetUserResponse, resError error)
	}{
		{
			testName: "user obtained",
			repoResponse: func(userId string) (repository.User, error) {
				return repository.User{
					UserId:  userId,
					Name:    user.Name,
					Age:     user.Age,
					AddInfo: user.AddInfo,
				}, nil
			},
			checkResponse: func(t *testing.T, res GetUserResponse, resError error) {
				assert.Equal(t, user.UserId, res.UserId)
				assert.Equal(t, user.Name, res.Name)
				assert.Equal(t, user.Age, res.Age)
				assert.Equal(t, user.AddInfo, res.AddInfo)
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user id empty",
			repoResponse: func(userId string) (repository.User, error) {
				return repository.User{}, nil
			},
			checkResponse: func(t *testing.T, res GetUserResponse, resError error) {
				assert.Empty(t, res)
				assert.EqualError(t, resError, erro.ErrRequiredFields("userId"))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			user, err := tc.repoResponse(user.UserId)
			repoSvc.On("GetUser", ctx, user.UserId).
				Return(user, err)
			res, err := service.GetUser(ctx, user.UserId)
			tc.checkResponse(t, res, err)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "service_test",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	repoSvc := new(repoMock)

	service := NewService(repoSvc, logger)

	testCases := []struct {
		testName string
		userData struct {
			UserId  string
			Name    string
			Pwd     string
			Age     uint32
			AddInfo string
		}
		userId        string
		request       func(userId, name, pwd, addInfo string, age uint32) UpdateUserRequest
		repoResponse  func() error
		checkResponse func(t *testing.T, resError error)
	}{
		{
			testName: "user updated",
			userData: struct {
				UserId  string
				Name    string
				Pwd     string
				Age     uint32
				AddInfo string
			}{
				utils.RandomString(12), "javier", "javier123", 45, "Some additional info"},
			request: func(userId, name, pwd, addInfo string, age uint32) UpdateUserRequest {
				return UpdateUserRequest{
					UserId:  userId,
					Name:    name,
					Pwd:     pwd,
					Age:     age,
					AddInfo: addInfo,
				}
			},
			repoResponse: func() error {
				return nil
			},
			checkResponse: func(t *testing.T, resError error) {
				assert.NoError(t, resError)
			},
		},
		{
			testName: "userId empty",
			userData: struct {
				UserId  string
				Name    string
				Pwd     string
				Age     uint32
				AddInfo string
			}{
				"", "javier", "javier123", 45, ""},
			request: func(userId, name, pwd, addInfo string, age uint32) UpdateUserRequest {
				return UpdateUserRequest{
					UserId:  userId,
					Name:    name,
					Pwd:     pwd,
					Age:     age,
					AddInfo: addInfo,
				}
			},
			repoResponse: nil,
			checkResponse: func(t *testing.T, resError error) {
				assert.EqualError(t, resError, erro.ErrRequiredFields("userId"))
			},
		},
		{
			testName: "no fields for update",
			userData: struct {
				UserId  string
				Name    string
				Pwd     string
				Age     uint32
				AddInfo string
			}{
				utils.RandomString(12), "", "", 0, ""},
			request: func(userId, name, pwd, addInfo string, age uint32) UpdateUserRequest {
				return UpdateUserRequest{
					UserId:  userId,
					Name:    name,
					Pwd:     pwd,
					Age:     age,
					AddInfo: addInfo,
				}
			},
			repoResponse: nil,
			checkResponse: func(t *testing.T, resError error) {
				assert.EqualError(t, resError, erro.ErrNoFieldsForUpdate)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.repoResponse != nil {
				err := tc.repoResponse()
				repoSvc.On("UpdateUser", ctx, repository.User{
					UserId:   tc.userData.UserId,
					Name:     tc.userData.Name,
					Password: tc.userData.Pwd,
					Age:      tc.userData.Age,
					AddInfo:  tc.userData.AddInfo,
				}).
					Return(err)
			}
			err := service.UpdateUser(ctx, tc.request(tc.userData.UserId, tc.userData.Name, tc.userData.Pwd, tc.userData.AddInfo, tc.userData.Age))
			tc.checkResponse(t, err)
		})
	}
}
