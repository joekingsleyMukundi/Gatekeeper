package db

import "context"

type CreateUserTxParam struct {
	CreateUserParams
	AfterCreate func(user User) error
}

type CreateUserTxResults struct {
	User User
}

func (store *SQLStorage) CreateUserTx(ctx context.Context, arg CreateUserTxParam) (CreateUserTxResults, error) {
	var result CreateUserTxResults
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}
		return arg.AfterCreate(result.User)
	})
	return result, err
}
