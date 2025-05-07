package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	CreateUserTx(ctx context.Context, arg CreateUserTxParam) (CreateUserTxResults, error)
	CreatePasswordResetTokenTx(ctx context.Context, arg CreatePasswordResetTokenTxParams) (CreatePasswordResetTokenTxResult, error)
}
type SQLStorage struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStorage{
		db:      db,
		Queries: New(db),
	}
}
func (store *SQLStorage) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	query := New(tx)
	err = fn(query)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}
