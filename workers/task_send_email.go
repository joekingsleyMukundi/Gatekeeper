package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/rs/zerolog/log"
)

const TaskSendEmail = "task:send_verfy_email"

type PayloadSendEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendEmail(
	ctx context.Context, payload *PayloadSendEmail, opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal send verify email paylosd to json: %w", err)
	}
	task := asynq.NewTask(TaskSendEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}
func (processor *RedisTaskPrcessor) ProcessTaskSendEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("Failed to unmashal payload in verify email task processor: %w", err)
	}
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		// if err == sql.ErrNoRows {
		// 	return fmt.Errorf("user doesnot exist: %w", asynq.SkipRetry)
		// }
		return fmt.Errorf("Failed to get user while processing verify email sender: %w", err)
	}
	emailverifyToken, err := utils.RandomByte(32)
	if err != nil {
		return fmt.Errorf("error creating email vrification token: %w", err)
	}
	emailVerifyTokenOHash := utils.HashRandomBytes(emailverifyToken)
	_, err = processor.store.CreateEmailVerifyToken(ctx, db.CreateEmailVerifyTokenParams{
		Username:  user.Username,
		Token:     emailVerifyTokenOHash,
		ExpiresAt: time.Now().Add(60 * time.Minute),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}
	subject := "Welcome to Gatekeeper"
	verifyUrl := fmt.Sprintf("http://localhost:8080/api/v1/auth/email/verify/%d", emailverifyToken)
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, user.Username, verifyUrl)
	to := []string{user.Email}
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %s", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
