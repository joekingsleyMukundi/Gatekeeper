package workers

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendEmail(
		ctx context.Context,
		payload *PayloadSendEmail,
		opts ...asynq.Option,
	) error
	DistributeTaskSendPasswordResetTokenEmail(
		ctx context.Context, payload *PayloadSendPasswordResetTokenEmail, opts ...asynq.Option,
	) error
}
type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOptions asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOptions)
	return &RedisTaskDistributor{
		client: client,
	}
}
