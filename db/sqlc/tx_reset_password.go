package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/joekingsleyMukundi/Gatekeeper/utils"
)

type TxResetPasswordParams struct {
	Token       string
	NewPassword string
	HashedToken string
	Now         time.Time
}

type TxResetPasswordResult struct {
	User User
}

func (store *SQLStorage) TxResetPassword(ctx context.Context, arg TxResetPasswordParams) (TxResetPasswordResult, error) {
	var result TxResetPasswordResult

	err := store.execTx(ctx, func(q *Queries) error {
		resetToken, err := q.GetActivePasswordResetToken(ctx, arg.HashedToken)
		if err != nil {
			return err
		}
		newHashedPassword, err := utils.HashPassword(arg.NewPassword)
		if err != nil {
			return err
		}
		arg := UpdateUserParams{
			HashedPassword: sql.NullString{
				String: newHashedPassword,
				Valid:  true,
			},
			PasswordChangedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			Username: resetToken.Owner,
		}
		user, err := q.UpdateUser(ctx, arg)
		if err != nil {
			return err
		}
		err = q.UpdatePasswordResetToken(ctx, resetToken.Token)
		if err != nil {
			return err
		}
		err = q.UpdatePasswordResetToken(ctx, resetToken.Owner)
		if err != nil {
			return err
		}
		result.User = user

		return nil
	})

	return result, err
}
