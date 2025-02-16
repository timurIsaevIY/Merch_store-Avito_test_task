package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"Merch_store-Avito_test_task/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name        string
	setupMock   func(mock sqlmock.Sqlmock)
	input       interface{}
	expectedRes interface{}
	expectedErr error
}

func TestAuthRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewAuthRepositoryImpl(db)

	testCases := []testCase{
		{
			name: "CreateUser - successful",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO "user" \(username, password_hash\) VALUES \(\$1, \$2\) RETURNING id`).
					WithArgs("test_user", "hashed_password").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			input: models.User{
				Username:     "test_user",
				PasswordHash: "hashed_password",
			},
			expectedRes: uint(1),
			expectedErr: nil,
		},
		{
			name: "CreateUser - user already exists",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO "user"`).
					WithArgs("test_user", "hashed_password").
					WillReturnError(&pq.Error{Code: "23505"}) // 23505 - дубликат
			},
			input: models.User{
				Username:     "test_user",
				PasswordHash: "hashed_password",
			},
			expectedRes: uint(0),
			expectedErr: models.ErrAlreadyExists,
		},
		{
			name: "CreateUser - database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO "user"`).
					WithArgs("test_user", "hashed_password").
					WillReturnError(errors.New("db error"))
			},
			input: models.User{
				Username:     "test_user",
				PasswordHash: "hashed_password",
			},
			expectedRes: uint(0),
			expectedErr: errors.New("db error"),
		},
		{
			name: "GetUser - successful",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, password_hash FROM "user" WHERE username = \$1`).
					WithArgs("test_user").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash"}).
						AddRow(1, "test_user", "hashed_password"))
			},
			input:       "test_user",
			expectedRes: models.User{ID: 1, Username: "test_user", PasswordHash: "hashed_password"},
			expectedErr: nil,
		},
		{
			name: "GetUser - user not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, password_hash FROM "user"`).
					WithArgs("unknown_user").
					WillReturnError(sql.ErrNoRows)
			},
			input:       "unknown_user",
			expectedRes: models.User{},
			expectedErr: models.ErrNotFound,
		},
		{
			name: "GetUser - database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, password_hash FROM "user"`).
					WithArgs("test_user").
					WillReturnError(errors.New("db error"))
			},
			input:       "test_user",
			expectedRes: models.User{},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock(mock)

			switch input := tc.input.(type) {
			case models.User:
				res, err := repo.CreateUser(context.Background(), input)
				assert.Equal(t, tc.expectedRes, res)
				assert.Equal(t, tc.expectedErr, err)

			case string:
				res, err := repo.GetUser(context.Background(), input)
				assert.Equal(t, tc.expectedRes, res)
				assert.Equal(t, tc.expectedErr, err)
			}
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
