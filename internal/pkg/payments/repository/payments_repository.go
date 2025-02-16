package repository

import (
	"Merch_store-Avito_test_task/internal/models"
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

type PaymentsRepositoryImpl struct {
	db *sql.DB
}

func NewAuthRepositoryImpl(db *sql.DB) *PaymentsRepositoryImpl {
	return &PaymentsRepositoryImpl{db}
}

func (r *PaymentsRepositoryImpl) Transfer(ctx context.Context, toUser string, amount uint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed start transaction: %w", err)
	}
	defer tx.Rollback()

	userID := ctx.Value(middleware.IdKey).(uint)

	query := `UPDATE "user" SET coins = coins - $1 WHERE id = $2`
	res, err := tx.ExecContext(ctx, query, amount, userID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23514" {
				return fmt.Errorf("not enough coins to send: %w", models.ErrNotEnough)
			}
		}
		return fmt.Errorf("updating balance failed: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected failed: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("receiver not found: %w", models.ErrNotFound)
	}

	query = `UPDATE "user" SET coins = coins + $1 WHERE username = $2`
	res, err = tx.ExecContext(ctx, query, amount, toUser)
	if err != nil {
		return fmt.Errorf("updating balance failed: %v", err)
	}
	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected failed: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("receiver not found: %w", models.ErrNotFound)
	}

	query = `INSERT INTO "transaction" (amount, from_user_id, to_user_id) VALUES ($1, $2, (SELECT id FROM "user" WHERE username = $3))`
	res, err = tx.ExecContext(ctx, query, amount, userID, toUser)
	if err != nil {
		return fmt.Errorf("inserting transaction failed: %v", err)
	}
	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected failed: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("receiver not found: %w", models.ErrNotFound)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction failed: %v", err)
	}
	return nil
}

func (r *PaymentsRepositoryImpl) BuyItem(ctx context.Context, itemID uint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed start transaction: %w", err)
	}
	defer tx.Rollback()
	userID := ctx.Value(middleware.IdKey).(uint)
	var amount uint
	query := `SELECT price FROM "product" WHERE id = $1`
	row := tx.QueryRowContext(ctx, query, itemID)
	err = row.Scan(&amount)
	if err != nil {
		return fmt.Errorf("getting product failed: %v", err)
	}
	query = `UPDATE "user" SET coins = coins - $1 WHERE id = $2`
	res, err := tx.ExecContext(ctx, query, amount, userID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23514" {
				return fmt.Errorf("not enough coins to buy: %w", models.ErrNotEnough)
			}
		}
		return fmt.Errorf("updating balance failed: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected failed: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("receiver not found: %w", models.ErrNotFound)
	}
	query = `INSERT INTO "purchase" (user_id, product_id) VALUES ($1, $2)`
	res, err = tx.ExecContext(ctx, query, userID, itemID)
	if err != nil {
		return fmt.Errorf("inserting purchase failed: %v", err)
	}
	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected failed: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("receiver not found: %w", models.ErrNotFound)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction failed: %v", err)
	}
	return nil
}
