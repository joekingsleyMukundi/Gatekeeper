package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendPasswordResetTokenEmail = "task:send_reset_password_email"

type PayloadSendPasswordResetTokenEmail struct {
	Username           string `json:"username"`
	ResetPasswordToken []byte `json:"reset_password_token"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendPasswordResetTokenEmail(
	ctx context.Context, payload *PayloadSendPasswordResetTokenEmail, opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal send password reset token payload to json: %w", err)
	}
	task := asynq.NewTask(TaskSendPasswordResetTokenEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue reset password task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}
func (processor *RedisTaskPrcessor) ProcessTaskSendPasswordResetTokenEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendPasswordResetTokenEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("Failed to unmashal payload in verify email task processor: %w", err)
	}
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user doesnot exist: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("Failed to get user while processing verify email sender: %w", err)
	}
	resetUrl := fmt.Sprintf("http://localhost:8080/api/v1/auth/password/reset/%s", payload.ResetPasswordToken)
	content := fmt.Sprintf("You are receiving this email because you (or someone else) has requested a reset of password. Please click this link, %s", resetUrl)
	subject := "Reset Password"
	to := []string{user.Email}
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %s", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}
