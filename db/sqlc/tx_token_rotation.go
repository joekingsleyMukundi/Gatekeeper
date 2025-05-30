package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joekingsleyMukundi/Gatekeeper/internals"
	"github.com/joekingsleyMukundi/Gatekeeper/tokens"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
)

type TxRenewTokenParams struct {
	RefreshToken        string
	RefreshTokenPayload *tokens.Payload
	TokenManager        internals.Manager
	TokenMaker          tokens.Maker
	Config              utils.Config
}
type TxRenewTokenResults struct {
	AccessToken         string
	AccessTokenPayload  *tokens.Payload
	RefreshToken        string
	RefreshTokenPayload *tokens.Payload
}

func (store *SQLStorage) TxRenewToken(ctx context.Context, arg TxRenewTokenParams) (TxRenewTokenResults, error) {
	var result TxRenewTokenResults
	var sessionToRevoke *Session
	err := store.execTx(ctx, func(q *Queries) error {
		sessionID, err := uuid.Parse(arg.RefreshTokenPayload.ID)
		if err != nil {
			return fmt.Errorf("invalid session ID format: %w", err)
		}
		session, err := q.GetSession(ctx, sessionID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("session not found")
			}
			return fmt.Errorf("failed to retrieve session: %w", err)
		}
		if session.IsBlocked {
			return fmt.Errorf("session is blocked")
		}

		if session.Username != arg.RefreshTokenPayload.Username {
			return fmt.Errorf("session username mismatch")
		}

		if session.RefreshToken != arg.RefreshToken {
			return fmt.Errorf("refresh token mismatch")
		}

		if time.Now().After(session.ExpiresAt) {
			return fmt.Errorf("refresh token expired")
		}
		isRevoked, _, err := arg.TokenManager.IsTokenRevoked(arg.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to check token revocation status: %w", err)
		}
		if isRevoked {
			sessionToRevoke = &session
			blockSessionArgs := UpdateSessionParams{
				ID: session.ID,
				IsBlocked: sql.NullBool{
					Bool:  true,
					Valid: true,
				},
			}
			if _, err := q.UpdateSession(ctx, blockSessionArgs); err != nil {
				return fmt.Errorf("failed to block compromised session: %w", err)
			}

			return fmt.Errorf("security violation: token reuse detected")
		}
		if err := arg.TokenManager.RevokeToken(arg.RefreshToken, arg.RefreshTokenPayload.ID); err != nil {
			return fmt.Errorf("failed to revoke old refresh token: %w", err)
		}
		accessToken, accessPayload, err := arg.TokenMaker.CreateToken(
			arg.RefreshTokenPayload.Username,
			arg.Config.AccessTokenDuration,
		)
		if err != nil {
			return fmt.Errorf("failed to create access token: %w", err)
		}
		refreshToken, refreshPayload, err := arg.TokenMaker.CreateToken(
			arg.RefreshTokenPayload.Username,
			arg.Config.RefreshTokenDuration,
		)
		if err != nil {
			return fmt.Errorf("failed to create refresh token: %w", err)
		}
		updateSessionArgs := UpdateSessionParams{
			ID: session.ID,
			RefreshToken: sql.NullString{
				Valid:  true,
				String: refreshToken,
			},
		}

		updatedSession, err := q.UpdateSession(ctx, updateSessionArgs)
		if err != nil {
			return fmt.Errorf("failed to update session: %w", err)
		}
		refreshPayload.ExpiresAt = jwt.NewNumericDate(updatedSession.ExpiresAt)

		result = TxRenewTokenResults{
			AccessToken:         accessToken,
			AccessTokenPayload:  accessPayload,
			RefreshToken:        refreshToken,
			RefreshTokenPayload: refreshPayload,
		}
		return nil
	})
	if sessionToRevoke != nil {
		if revokeErr := arg.TokenManager.RevokeToken(sessionToRevoke.RefreshToken, sessionToRevoke.ID.String()); revokeErr != nil {
			fmt.Printf("Warning: failed to revoke token for session %s: %v\n", sessionToRevoke.ID, revokeErr)
		}
	}
	return result, err
}
