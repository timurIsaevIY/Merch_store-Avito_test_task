package repository

import (
	"Merch_store-Avito_test_task/internal/models"
	"context"
	"database/sql"
	"fmt"
)

type ServiceRepoImpl struct {
	db *sql.DB
}

func NewServiceRepo(db *sql.DB) *ServiceRepoImpl {
	return &ServiceRepoImpl{db}
}

func (r *ServiceRepoImpl) GetUserInfo(ctx context.Context, userID uint) (models.UserData, error) {
	var data models.UserData
	query := `SELECT coins from "user" WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, userID)
	err := row.Scan(&data.Coins)
	if err != nil {
		return models.UserData{}, fmt.Errorf("failed to get balance: %w", err)
	}

	query = `SELECT p.name, COUNT(p.id)
		FROM "purchase" pu
		JOIN "product" p ON pu.product_id = p.id
		WHERE pu.user_id = $1
		GROUP BY p.name`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return models.UserData{}, fmt.Errorf("failed to get purchases: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var inventory models.Inventory
		if err = rows.Scan(&inventory.Type, &inventory.Quantity); err != nil {
			return models.UserData{}, fmt.Errorf("failed to scan inventory: %w", err)
		}
		data.Inventory = append(data.Inventory, inventory)
	}

	query = `SELECT u.username, t.amount
		FROM transaction t
		JOIN "user" u ON t.from_user_id = u.id
		WHERE t.to_user_id = $1 ORDER BY t.created_at DESC`
	rows, err = r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return models.UserData{}, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var transaction models.Transaction
		err = rows.Scan(&transaction.FromUser, &transaction.Amount)
		if err != nil {
			return models.UserData{}, fmt.Errorf("failed to scan transactions: %w", err)
		}
		data.CoinHistory.Received = append(data.CoinHistory.Received, transaction)
	}
	query = `SELECT u.username, t.amount
		FROM transaction t
		JOIN "user" u ON t.to_user_id = u.id
		WHERE t.from_user_id = $1 ORDER BY t.created_at DESC`
	rows, err = r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return models.UserData{}, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var transaction models.Transaction
		err = rows.Scan(&transaction.ToUser, &transaction.Amount)
		if err != nil {
			return models.UserData{}, fmt.Errorf("failed to scan transactions: %w", err)
		}
		data.CoinHistory.Sent = append(data.CoinHistory.Sent, transaction)
	}
	return data, nil
}
