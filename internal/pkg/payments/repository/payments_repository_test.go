package repository

import (
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	"context"
	"database/sql"
	"errors"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPaymentsRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	repo := &PaymentsRepositoryImpl{db}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"Transfer - Successful", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectExec(`UPDATE "user" SET coins = coins - \$1 WHERE id = \$2`).
				WithArgs(100, 1).
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectExec(`UPDATE "user" SET coins = coins \+ \$1 WHERE username = \$2`).
				WithArgs(100, "receiver").
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectExec(`INSERT INTO "transaction" \(amount, from_user_id, to_user_id\) VALUES \(\$1, \$2, \(SELECT id FROM "user" WHERE username = \$3\)\)`).
				WithArgs(100, 1, "receiver").
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectCommit()

			err := repo.Transfer(context.WithValue(context.Background(), middleware.IdKey, uint(1)), "receiver", 100)
			assert.NoError(t, err)
		}},

		{"Transfer - Not Enough Coins", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE "user" SET coins = coins - \$1 WHERE id = \$2`).
				WithArgs(100, 1).
				WillReturnError(errors.New("not enough coins to send"))
			mock.ExpectRollback()

			err := repo.Transfer(context.WithValue(context.Background(), middleware.IdKey, uint(1)), "receiver", 100)
			assert.Error(t, err)
		}},

		{"Transfer - Receiver Not Found", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE "user" SET coins = coins - \$1 WHERE id = \$2`).
				WithArgs(100, 1).
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectExec(`UPDATE "user" SET coins = coins \+ \$1 WHERE username = \$2`).
				WithArgs(100, "unknown_user").
				WillReturnResult(sqlmock.NewResult(0, 0))

			mock.ExpectRollback()

			err := repo.Transfer(context.WithValue(context.Background(), middleware.IdKey, uint(1)), "unknown_user", 100)
			assert.Error(t, err)
		}},

		{"BuyItem - Successful", func(t *testing.T) {
			mock.ExpectBegin()

			mock.ExpectQuery(`SELECT price FROM "product" WHERE id = \$1`).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(500))

			mock.ExpectExec(`UPDATE "user" SET coins = coins - \$1 WHERE id = \$2`).
				WithArgs(500, 1).
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectExec(`INSERT INTO "purchase" \(user_id, product_id\) VALUES \(\$1, \$2\)`).
				WithArgs(1, 1).
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectCommit()

			err := repo.BuyItem(context.WithValue(context.Background(), middleware.IdKey, uint(1)), 1)
			assert.NoError(t, err)
		}},

		{"BuyItem - Not Enough Coins", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT price FROM "product" WHERE id = \$1`).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(500))

			mock.ExpectExec(`UPDATE "user" SET coins = coins - \$1 WHERE id = \$2`).
				WithArgs(500, 1).
				WillReturnError(errors.New("not enough coins to buy"))

			mock.ExpectRollback()

			err := repo.BuyItem(context.WithValue(context.Background(), middleware.IdKey, uint(1)), 1)
			assert.Error(t, err)
		}},

		{"BuyItem - Product Not Found", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT price FROM "product" WHERE id = \$1`).
				WithArgs(999).
				WillReturnError(sql.ErrNoRows)

			mock.ExpectRollback()

			err := repo.BuyItem(context.WithValue(context.Background(), middleware.IdKey, uint(1)), 999)
			assert.Error(t, err)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
