package service

import (
	"context"
	"os"
	"testing"

	"github.com/go-kit/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/javibauza/final-project/grpc-service/entities"
	erro "github.com/javibauza/final-project/grpc-service/errors"
	"github.com/javibauza/final-project/grpc-service/utils"
)

type repoMock struct {
	mock.Mock
}

func (m *repoMock) Authenticate(ctx context.Context, name string) (entities.User, error) {
	args := m.Called(ctx, name)

	if args.Get(0) == nil {
		return entities.User{}, args.Error(1)
	}

	return args.Get(0).(entities.User), args.Error(1)
}

func (m *repoMock) CreateUser(ctx context.Context, user entities.User) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *repoMock) UpdateUser(ctx context.Context, user entities.User) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *repoMock) GetUser(ctx context.Context, userId string) (entities.User, error) {
	args := m.Called(ctx, userId)

	return args.Get(0).(entities.User), args.Error(1)
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
		request       func(name, pwd string) entities.User
		repoResponse  func(userId string, pwdHash []byte) (entities.User, error)
		checkResponse func(t *testing.T, response, userId string, resError error)
	}{
		{
			testName:    "user authenticated",
			userName:    "javier",
			userPwd:     "javier123",
			userId:      utils.RandomString(12),
			userPwdHash: []byte("$2a$12$RXSLrffQZDUGljSPdQAPI.W4txkPKkeASl0qSM/tbx7mgMMqnDhui"),
			request: func(name, pwd string) entities.User {
				return entities.User{Name: name, Password: pwd}
			},
			repoResponse: func(userId string, pwdHash []byte) (entities.User, error) {
				return entities.User{UserId: userId, PwdHash: string(pwdHash)}, nil
			},
			checkResponse: func(t *testing.T, response, userId string, resError error) {
				assert.Equal(t, response, userId)
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user not found",
			userName: "reivaj",
			userPwd:  "javier123",
			request: func(name, pwd string) entities.User {
				return entities.User{Name: name, Password: pwd}
			},
			repoResponse: func(userId string, pwdHash []byte) (entities.User, error) {
				return entities.User{}, erro.NewErrNotFound()
			},
			checkResponse: func(t *testing.T, response, userId string, resError error) {
				assert.Empty(t, response)
				_, ok := resError.(*erro.ErrNotFound)
				assert.EqualValues(t, true, ok)
			},
		},
		{
			testName:    "wrong password",
			userName:    "javier",
			userPwd:     "javier321",
			userId:      utils.RandomString(12),
			userPwdHash: []byte("$2a$12$RXSLrffQZDUGljSPdQAPI.W4txkPKkeASl0qSM/tbx7mgMMqnDhui"),
			request: func(name, pwd string) entities.User {
				return entities.User{Name: name, Password: pwd}
			},
			repoResponse: func(userId string, pwdHash []byte) (entities.User, error) {
				return entities.User{UserId: userId, PwdHash: string(pwdHash)}, nil
			},
			checkResponse: func(t *testing.T, response, userId string, resError error) {
				assert.Empty(t, response)
				res, ok := resError.(*erro.ErrPermissionDenied)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrWrongPassword)
			},
		},
		{
			testName:    "user name empty",
			userName:    "",
			userPwd:     "javier321",
			userId:      utils.RandomString(12),
			userPwdHash: []byte("$2a$12$RXSLrffQZDUGljSPdQAPI.W4txkPKkeASl0qSM/tbx7mgMMqnDhui"),
			request: func(name, pwd string) entities.User {
				return entities.User{Name: name, Password: pwd}
			},
			repoResponse: nil,
			checkResponse: func(t *testing.T, response, userId string, resError error) {
				assert.Empty(t, response)
				res, ok := resError.(*erro.ErrInvalidArgument)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrRequiredFields("name", "password"))
			},
		},
		{
			testName:    "password empty",
			userName:    "javier",
			userPwd:     "",
			userId:      utils.RandomString(12),
			userPwdHash: []byte("$2a$12$RXSLrffQZDUGljSPdQAPI.W4txkPKkeASl0qSM/tbx7mgMMqnDhui"),
			request: func(name, pwd string) entities.User {
				return entities.User{Name: name, Password: pwd}
			},
			repoResponse: nil,
			checkResponse: func(t *testing.T, response, userId string, resError error) {
				assert.Empty(t, response)
				res, ok := resError.(*erro.ErrInvalidArgument)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrRequiredFields("name", "password"))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			if tc.repoResponse != nil {
				repoResponse, err := tc.repoResponse(tc.userId, tc.userPwdHash)
				repoSvc.On("Authenticate", ctx, tc.userName).
					Return(repoResponse, err)
			}

			res, err := service.Authenticate(ctx, tc.request(tc.userName, tc.userPwd))
			tc.checkResponse(t, res, tc.userId, err)
		})
	}
}

/*
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
		pwdHash       string
		request       func(name, pwd, addInfo string, age uint32) CreateUserRequest
		repoResponse  func(userId string) error
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
			pwdHash: "$2a$12$RXSLrffQZDUGljSPdQAPI.W4txkPKkeASl0qSM/tbx7mgMMqnDhui",
			userId:  utils.RandomString(12),
			request: func(name, pwd, addInfo string, age uint32) CreateUserRequest {
				return CreateUserRequest{
					Name:    name,
					Pwd:     pwd,
					Age:     age,
					AddInfo: addInfo,
				}
			},
			repoResponse: func(userId string) error {
				return nil
			},
			checkResponse: func(t *testing.T, userId string, response CreateUserResponse, resError error) {
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
			repoResponse: func(userId string) error {
				return nil
			},
			checkResponse: func(t *testing.T, userId string, response CreateUserResponse, resError error) {
				assert.Empty(t, response)
				res, ok := resError.(*erro.ErrInvalidArgument)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrRequiredFields("name", "password"))
			},
		},
		{
			testName: "password empty",
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
			repoResponse: func(userId string) error {
				return nil
			},
			checkResponse: func(t *testing.T, userId string, response CreateUserResponse, resError error) {
				assert.Empty(t, response)
				assert.Empty(t, response)
				res, ok := resError.(*erro.ErrInvalidArgument)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrRequiredFields("name", "password"))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			repoErr := tc.repoResponse(tc.userId)
			repoSvc.On("CreateUser", ctx, mock.AnythingOfType("entities.User")).
				Return(repoErr)
			res, err := service.CreateUser(ctx, tc.request(tc.userData.Name, tc.userData.Pwd, tc.userData.AddInfo, tc.userData.Age))
			tc.checkResponse(t, tc.userId, res, err)
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
		request       func(userId, name, pwd, addInfo string, age uint32) UpdateUserRequest
		repoResponse  error
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
			repoResponse: nil,
			checkResponse: func(t *testing.T, resError error) {
				assert.NoError(t, resError)
			},
		},
		{
			testName: "empty userId",
			userData: struct {
				UserId  string
				Name    string
				Pwd     string
				Age     uint32
				AddInfo string
			}{
				"", "javier", "javier123", 45, "Some additional info"},
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
				res, ok := resError.(*erro.ErrInvalidArgument)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrRequiredFields("userId"))
			},
		},
		{
			testName: "no fields por update",
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
				res, ok := resError.(*erro.ErrInvalidArgument)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrNoFieldsForUpdate)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			repoSvc.On("UpdateUser", ctx, mock.AnythingOfType("entities.User")).
				Return(tc.repoResponse)
			err := service.UpdateUser(ctx, tc.request(tc.userData.UserId, tc.userData.Name, tc.userData.Pwd, tc.userData.AddInfo, tc.userData.Age))
			tc.checkResponse(t, err)
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

	testCases := []struct {
		testName      string
		userId        string
		repoResponse  func(userId string) (entities.User, error)
		checkResponse func(t *testing.T, userId string, response GetUserResponse, resError error)
	}{
		{
			testName: "user obtained",
			userId:   utils.RandomString(12),
			repoResponse: func(userId string) (entities.User, error) {
				return entities.User{
					UserId:  userId,
					Age:     45,
					Name:    "javier",
					AddInfo: sql.NullString{String: "additional info"},
				}, nil
			},
			checkResponse: func(t *testing.T, userId string, response GetUserResponse, resError error) {
				assert.Equal(t, response.UserId, userId)
				assert.NoError(t, resError)
			},
		},
		{
			testName: "empty userId",
			userId:   "",
			repoResponse: func(userId string) (entities.User, error) {
				return entities.User{}, nil
			},
			checkResponse: func(t *testing.T, userId string, response GetUserResponse, resError error) {
				res, ok := resError.(*erro.ErrInvalidArgument)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrRequiredFields("userId"))
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			repoResponse, err := tc.repoResponse(tc.userId)
			repoSvc.On("GetUser", ctx, tc.userId).
				Return(repoResponse, err)
			user, err := service.GetUser(ctx, tc.userId)
			tc.checkResponse(t, tc.userId, user, err)
		})
	}
}
*/
