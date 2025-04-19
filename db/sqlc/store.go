package db

import "database/sql"

type Store interface {
	Querier
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
