package db

import "context"

type CreatePasswordResetTokenTxParams struct {
	CreatePasswordResetTokenParams CreatePasswordResetTokenParams
	AfterCreate                    func(passwordResetToken PasswordResetToken) error
}

type CreatePasswordResetTokenTxResult struct {
	PasswordResetToken PasswordResetToken
}

func (store *SQLStorage) CreatePasswordResetTokenTx(ctx context.Context, arg CreatePasswordResetTokenTxParams) (CreatePasswordResetTokenTxResult, error) {
	var result CreatePasswordResetTokenTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.PasswordResetToken, err = q.CreatePasswordResetToken(ctx, arg.CreatePasswordResetTokenParams)
		if err != nil {
			return err
		}
		return arg.AfterCreate(result.PasswordResetToken)
	})
	return result, err
}
