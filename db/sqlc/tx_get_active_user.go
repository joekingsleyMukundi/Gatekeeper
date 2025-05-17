package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joekingsleyMukundi/Gatekeeper/tokens"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
)

type TxLoginUserParams struct {
	Username             string
	Password             string
	UserAgent            string
	ClientIP             string
	TokenMaker           tokens.Maker
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type TxLoginUserResult struct {
	User           User
	Session        Session
	AccessToken    string
	AccessPayload  *tokens.Payload
	RefreshToken   string
	RefreshPayload *tokens.Payload
}

func (store *SQLStorage) TxLoginUser(ctx context.Context, arg TxLoginUserParams) (TxLoginUserResult, error) {
	var result TxLoginUserResult
	err := store.execTx(ctx, func(q *Queries) error {
		user, err := q.GetUser(ctx, arg.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("user not found")
			}
			return err
		}

		verified, err := q.IsUserEmailVerified(ctx, user.Username)
		if err != nil {
			return err
		}
		if !verified {
			return fmt.Errorf("email not verified")
		}

		err = utils.ConfirmPassword(arg.Password, user.HashedPassword)
		if err != nil {
			return fmt.Errorf("invalid credentials")
		}
		accessToken, accessPayload, err := arg.TokenMaker.CreateToken(user.Username, arg.AccessTokenDuration)
		if err != nil {
			return err
		}
		refreshToken, refreshPayload, err := arg.TokenMaker.CreateToken(user.Username, arg.RefreshTokenDuration)
		if err != nil {
			return err
		}
		refreshPayloadId, err := uuid.Parse(refreshPayload.ID)
		if err != nil {
			return fmt.Errorf("Error parsing uuid: %d", err)
		}
		session, err := q.CreateSession(ctx, CreateSessionParams{
			ID:           refreshPayloadId,
			Username:     user.Username,
			RefreshToken: refreshToken,
			UserAgent:    arg.UserAgent,
			ClientIp:     arg.ClientIP,
			IsBlocked:    false,
			ExpiresAt:    refreshPayload.ExpiresAt.Time,
		})
		if err != nil {
			return err
		}
		result = TxLoginUserResult{
			User:           user,
			Session:        session,
			AccessToken:    accessToken,
			AccessPayload:  accessPayload,
			RefreshToken:   refreshToken,
			RefreshPayload: refreshPayload,
		}
		return nil
	})
	return result, err
}
