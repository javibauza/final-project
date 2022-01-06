package repository

import (
	"context"
	"database/sql"

	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/assert"

	"github.com/javibauza/final-project/grpc-service/entities"
	erro "github.com/javibauza/final-project/grpc-service/errors"
	"github.com/javibauza/final-project/grpc-service/utils"
)

var user = &entities.User{
	Id:      001,
	UserId:  utils.RandomString(12),
	PwdHash: "$2a$12$Lhdc.gbeLbQbm8uz3H1T4.EPaxqclyblPeM1N1rxhNCth1/sZkCwC", //jbauza123
	Name:    "javier",
	Age:     37,
	AddInfo: sql.NullString{},
}

func NewMock(logger log.Logger) (*sql.DB, sqlmock.Sqlmock) {

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		level.Error(logger).Log("error opening a stub database connection", err)
	}

	return db, mock
}

func TestAuthenticate(t *testing.T) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "repo_test",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	db, mock := NewMock(logger)
	defer db.Close()

	repo := NewRepo(db, logger)

	testCases := []struct {
		testName      string
		userName      string
		buildStubs    func(mock sqlmock.Sqlmock, userName string)
		checkResponse func(t *testing.T, response entities.User, resError error)
	}{
		{
			testName: "user authenticated",
			userName: "javier",
			buildStubs: func(mock sqlmock.Sqlmock, userName string) {
				rows := sqlmock.NewRows([]string{"user_id", "pwd_hash"}).AddRow(user.UserId, user.PwdHash)
				mock.ExpectPrepare(authenticateSQL)
				mock.ExpectQuery(authenticateSQL).WithArgs(userName).WillReturnRows(rows)
			},
			checkResponse: func(t *testing.T, response entities.User, resError error) {
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user not found",
			userName: "reivaj",
			buildStubs: func(mock sqlmock.Sqlmock, userName string) {
				rows := sqlmock.NewRows([]string{"user_id", "pwd_hash"})
				mock.ExpectPrepare(authenticateSQL)
				mock.ExpectQuery(authenticateSQL).WithArgs(userName).WillReturnRows(rows)
			},
			checkResponse: func(t *testing.T, response entities.User, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, erro.ErrUserNotFound)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()

			tc.buildStubs(mock, tc.userName)

			res, err := repo.Authenticate(ctx, tc.userName)
			tc.checkResponse(t, res, err)
		})
	}
}

func TestCreateUser(t *testing.T) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "repo_test",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	db, mock := NewMock(logger)
	defer db.Close()

	repo := NewRepo(db, logger)

	testCases := []struct {
		testName      string
		userData      *entities.User
		buildStubs    func(mock sqlmock.Sqlmock, user *entities.User)
		checkResponse func(t *testing.T, err error)
	}{
		{
			testName: "user created",
			userData: &entities.User{
				UserId:  user.UserId,
				PwdHash: user.PwdHash,
				Name:    user.Name,
				Age:     user.Age,
				AddInfo: user.AddInfo,
			},
			buildStubs: func(mock sqlmock.Sqlmock, request *entities.User) {
				var lastInsertID, affected int64
				mock.ExpectPrepare(createSQL)
				mock.ExpectExec(createSQL).
					WithArgs(request.UserId, request.Name, request.PwdHash, request.Age, request.AddInfo).
					WillReturnResult(sqlmock.NewResult(lastInsertID, affected))
			},
			checkResponse: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()

			tc.buildStubs(mock, tc.userData)

			request := entities.User{
				UserId:  tc.userData.UserId,
				PwdHash: tc.userData.PwdHash,
				Name:    tc.userData.Name,
				Age:     tc.userData.Age,
				AddInfo: tc.userData.AddInfo,
			}
			err := repo.CreateUser(ctx, request)
			tc.checkResponse(t, err)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger,
			"service", "repo_test",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	db, mock := NewMock(logger)
	defer db.Close()

	repo := NewRepo(db, logger)

	testCases := []struct {
		testName      string
		userData      *entities.User
		buildStubs    func(mock sqlmock.Sqlmock, user *entities.User)
		checkResponse func(t *testing.T, resError error)
	}{
		{
			testName: "user updated",
			userData: &entities.User{
				PwdHash: user.PwdHash,
				Name:    user.Name,
				Age:     user.Age,
			},
			buildStubs: func(mock sqlmock.Sqlmock, user *entities.User) {
				_, query := updateSQL(user)
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(user.PwdHash, user.Age, user.Name, user.UserId).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			checkResponse: func(t *testing.T, resError error) {
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user not found",
			userData: &entities.User{
				Name:    user.Name,
				PwdHash: user.PwdHash,
				Age:     user.Age,
			},
			buildStubs: func(mock sqlmock.Sqlmock, user *entities.User) {
				_, query := updateSQL(user)
				mock.ExpectPrepare(query)
				mock.ExpectExec(query).
					WithArgs(user.PwdHash, user.Age, user.Name, user.UserId).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			checkResponse: func(t *testing.T, resError error) {
				res, ok := resError.(*erro.ErrNotFound)
				assert.EqualValues(t, true, ok)
				assert.Equal(t, res.Err.Error(), erro.ErrUserNotFound)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()

			tc.buildStubs(mock, tc.userData)

			request := entities.User{
				UserId:  tc.userData.UserId,
				Name:    tc.userData.Name,
				PwdHash: tc.userData.PwdHash,
				Age:     tc.userData.Age,
				AddInfo: tc.userData.AddInfo,
			}
			err := repo.UpdateUser(ctx, request)
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
			"service", "repo_test",
			"time:", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	db, mock := NewMock(logger)
	defer db.Close()

	repo := NewRepo(db, logger)

	testCases := []struct {
		testName      string
		userId        string
		buildStubs    func(mock sqlmock.Sqlmock, userId string)
		checkResponse func(t *testing.T, response entities.User, resError error)
	}{
		{
			testName: "user obtained",
			userId:   "",
			buildStubs: func(mock sqlmock.Sqlmock, userId string) {
				rows := sqlmock.NewRows([]string{"user_id", "name", "age", "additional_information"}).
					AddRow(user.UserId, user.Name, user.Age, user.AddInfo)

				mock.ExpectPrepare(getSQL)
				mock.ExpectQuery(getSQL).
					WithArgs(userId).
					WillReturnRows(rows)
			},
			checkResponse: func(t *testing.T, response entities.User, resError error) {
				assert.NoError(t, resError)
			},
		},
		{
			testName: "user not found",
			userId:   "",
			buildStubs: func(mock sqlmock.Sqlmock, userId string) {
				rows := sqlmock.NewRows([]string{"user_id", "pwd_hash"})
				mock.ExpectPrepare(getSQL)
				mock.ExpectQuery(getSQL).
					WithArgs(userId).WillReturnRows(rows)
			},
			checkResponse: func(t *testing.T, response entities.User, resError error) {
				assert.Empty(t, response)
				assert.EqualError(t, resError, erro.ErrUserNotFound)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()

			tc.buildStubs(mock, tc.userId)

			res, err := repo.GetUser(ctx, tc.userId)
			tc.checkResponse(t, res, err)
		})
	}
}
