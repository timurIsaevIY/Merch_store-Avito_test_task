package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetUserInfo(t *testing.T) {
	// Создаём мок базы данных
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewServiceRepo(db)
	userID := uint(1)

	t.Run("Successful user info retrieval", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectQuery(`SELECT coins from "user" WHERE id = \$1`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(1000))

		mock.ExpectQuery(`SELECT p.name, COUNT\(p.id\) FROM "purchase" pu JOIN "product" p ON pu.product_id = p.id WHERE pu.user_id = \$1 GROUP BY p.name`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"name", "count"}).
				AddRow("Sword", 2).
				AddRow("Shield", 1))

		mock.ExpectQuery(`SELECT u.username, t.amount FROM transaction t JOIN "user" u ON t.from_user_id = u.id WHERE t.to_user_id = \$1 ORDER BY t.created_at DESC`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"username", "amount"}).
				AddRow("Alice", 500).
				AddRow("Bob", 300))

		mock.ExpectQuery(`SELECT u.username, t.amount FROM transaction t JOIN "user" u ON t.to_user_id = u.id WHERE t.from_user_id = \$1 ORDER BY t.created_at DESC`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"username", "amount"}).
				AddRow("Charlie", 200).
				AddRow("David", 100))

		data, err := repo.GetUserInfo(ctx, userID)

		assert.NoError(t, err)

		assert.Equal(t, 1000, data.Coins)

		assert.Len(t, data.Inventory, 2)
		assert.Equal(t, "Sword", data.Inventory[0].Type)
		assert.Equal(t, 2, data.Inventory[0].Quantity)
		assert.Equal(t, "Shield", data.Inventory[1].Type)
		assert.Equal(t, 1, data.Inventory[1].Quantity)

		assert.Len(t, data.CoinHistory.Received, 2)
		assert.Equal(t, "Alice", data.CoinHistory.Received[0].FromUser)
		assert.Equal(t, 500, data.CoinHistory.Received[0].Amount)
		assert.Equal(t, "Bob", data.CoinHistory.Received[1].FromUser)
		assert.Equal(t, 300, data.CoinHistory.Received[1].Amount)

		assert.Len(t, data.CoinHistory.Sent, 2)
		assert.Equal(t, "Charlie", data.CoinHistory.Sent[0].ToUser)
		assert.Equal(t, 200, data.CoinHistory.Sent[0].Amount)
		assert.Equal(t, "David", data.CoinHistory.Sent[1].ToUser)
		assert.Equal(t, 100, data.CoinHistory.Sent[1].Amount)
	})

	t.Run("User not found", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectQuery(`SELECT coins from "user" WHERE id = \$1`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		data, err := repo.GetUserInfo(ctx, userID)

		assert.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("Database error on balance retrieval", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectQuery(`SELECT coins from "user" WHERE id = \$1`).
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		data, err := repo.GetUserInfo(ctx, userID)

		assert.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("Database error on purchase retrieval", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectQuery(`SELECT coins from "user" WHERE id = \$1`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(1000))

		mock.ExpectQuery(`SELECT p.name, COUNT\(p.id\) FROM "purchase" pu JOIN "product" p ON pu.product_id = p.id WHERE pu.user_id = \$1 GROUP BY p.name`).
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		data, err := repo.GetUserInfo(ctx, userID)

		assert.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("Database error on transaction retrieval", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectQuery(`SELECT coins from "user" WHERE id = \$1`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(1000))

		mock.ExpectQuery(`SELECT p.name, COUNT\(p.id\) FROM "purchase" pu JOIN "product" p ON pu.product_id = p.id WHERE pu.user_id = \$1 GROUP BY p.name`).
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"name", "count"}).AddRow("Sword", 2))

		mock.ExpectQuery(`SELECT u.username, t.amount FROM transaction t JOIN "user" u ON t.from_user_id = u.id WHERE t.to_user_id = \$1 ORDER BY t.created_at DESC`).
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		data, err := repo.GetUserInfo(ctx, userID)

		assert.Error(t, err)
		assert.Empty(t, data)
	})
}
