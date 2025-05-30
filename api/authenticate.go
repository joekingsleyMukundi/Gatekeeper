package api

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/joekingsleyMukundi/Gatekeeper/workers"
	"github.com/lib/pq"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}
type createUserResponse struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(user db.User) createUserResponse {
	return createUserResponse{
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := db.CreateUserTxParam{
		CreateUserParams: db.CreateUserParams{
			Username:       req.Username,
			Email:          req.Email,
			HashedPassword: hashedPassword,
		},
		AfterCreate: func(user db.User) error {
			taskPayload := &workers.PayloadSendEmail{
				Username: req.Username,
			}
			opt := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(workers.QueueCrtical),
			}
			return server.taskDistributor.DistributeTaskSendEmail(ctx, taskPayload, opt...)
		},
	}
	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	resp := newUserResponse(txResult.User)
	ctx.JSON(http.StatusOK, resp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required"`
}
type loginUserResponse struct {
	SessionID             uuid.UUID          `json:"session_id"`
	AccessToken           string             `json:"access_token"`
	AccessTokenExpiresAt  time.Time          `json:"access_token_expires_at"`
	RefreshToken          string             `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time          `json:"refresh_token_expires_at"`
	User                  createUserResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	txResult, err := server.store.TxLoginUser(ctx, db.TxLoginUserParams{
		Username:             req.Username,
		Password:             req.Password,
		UserAgent:            ctx.Request.UserAgent(),
		ClientIP:             ctx.ClientIP(),
		TokenMaker:           server.TokenMaker,
		AccessTokenDuration:  server.config.AccessTokenDuration,
		RefreshTokenDuration: server.config.RefreshTokenDuration,
	})
	if err != nil {
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		if err.Error() == "email not verified" || err.Error() == "invalid credentials" {
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		SessionID:             txResult.Session.ID,
		AccessToken:           txResult.AccessToken,
		AccessTokenExpiresAt:  txResult.AccessPayload.ExpiresAt.Time,
		RefreshToken:          txResult.RefreshToken,
		RefreshTokenExpiresAt: txResult.RefreshPayload.ExpiresAt.Time,
		User:                  newUserResponse(txResult.User),
	}
	/**
	  Using SameSite and Secure cookie attributes
	  	http.SetCookie(ctx.Writer, &http.Cookie{
	  		Name:     "access_token",
	  		Value:    txResult.AccessToken,
	  		MaxAge:   900,
	  		Path:     "/",
	  		Secure:   true,
	  		HttpOnly: true,
	  		SameSite: http.SameSiteStrictMode,
	  	})

	  	http.SetCookie(ctx.Writer, &http.Cookie{
	  		Name:     "refresh_token",
	  		Value:    txResult.RefreshToken,
	  		MaxAge:   604800,
	  		Path:     "/auth/refresh",
	  		Secure:   true,
	  		HttpOnly: true,
	  		SameSite: http.SameSiteStrictMode,
	  	})
	  **/
	ctx.JSON(http.StatusOK, rsp)
}

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	refreshTokenPayload, err := server.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	txParams := db.TxRenewTokenParams{
		RefreshToken:        req.RefreshToken,
		RefreshTokenPayload: refreshTokenPayload,
		TokenManager:        server.tokenManager,
		TokenMaker:          server.TokenMaker,
		Config:              server.config,
	}
	result, err := server.store.TxRenewToken(ctx, txParams)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "session not found"):
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		case strings.Contains(err.Error(), "session is blocked"),
			strings.Contains(err.Error(), "session username mismatch"),
			strings.Contains(err.Error(), "refresh token mismatch"),
			strings.Contains(err.Error(), "refresh token expired"),
			strings.Contains(err.Error(), "security violation"):
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		default:
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}
	rsp := renewAccessTokenResponse{
		AccessToken:          result.AccessToken,
		AccessTokenExpiresAt: result.AccessTokenPayload.ExpiresAt.Time,
	}
	ctx.JSON(http.StatusOK, rsp)
}

type forgortPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}
type forgotPasswordResponse struct {
	Message string `json:"message"`
}

func (server *Server) forgotPassword(ctx *gin.Context) {
	var req forgortPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	user, err := server.store.LocateUser(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	resetToken := utils.RandomString(20)
	psResetTokenHash := utils.HashRandomBytes([]byte(resetToken))
	arg := db.CreatePasswordResetTokenTxParams{
		CreatePasswordResetTokenParams: db.CreatePasswordResetTokenParams{
			Owner:     user.Username,
			Token:     psResetTokenHash,
			ExpiresAt: time.Now().Add(server.config.PasswordResetTokenDuration),
		},
		AfterCreate: func(passwordResetToken db.PasswordResetToken) error {
			taskPayload := &workers.PayloadSendPasswordResetTokenEmail{
				Username:           user.Username,
				ResetPasswordToken: resetToken,
			}
			opt := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(workers.QueueCrtical),
			}
			return server.taskDistributor.DistributeTaskSendPasswordResetTokenEmail(ctx, taskPayload, opt...)
		},
	}
	_, err = server.store.CreatePasswordResetTokenTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	rsp := forgotPasswordResponse{
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, rsp)
}

type resetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}
type resetPasswordResponse struct {
	Message string `json:"message"`
}

func (server *Server) resetPassword(ctx *gin.Context) {
	psResetTokenHash := utils.HashRandomBytes([]byte(ctx.Param("token")))
	var req resetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, err := server.store.TxResetPassword(ctx, db.TxResetPasswordParams{
		Token:       ctx.Param("token"),
		NewPassword: req.Password,
		HashedToken: psResetTokenHash,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}
	rsp := resetPasswordResponse{
		Message: "Password reset successful",
	}
	ctx.JSON(http.StatusOK, rsp)
}

type emailValidationResponse struct {
	Message string `json:"message"`
}

func (server *Server) validateEmail(ctx *gin.Context) {
	emailVerificationTokenHash := utils.HashRandomBytes([]byte(ctx.Param("token")))
	emailVerificationTokenData, err := server.store.GetActiveEmailVerifyToken(ctx, emailVerificationTokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	err = server.store.UpdateEmailVerifyToken(ctx, emailVerificationTokenData.Token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	rsp := emailValidationResponse{
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, rsp)
}
