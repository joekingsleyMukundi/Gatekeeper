package workers

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/services/mail"
	"github.com/rs/zerolog/log"
)

const (
	QueueCrtical = "critical"
	QueueDefault = "default"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendEmail(ctx context.Context, task *asynq.Task) error
}
type RedisTaskPrcessor struct {
	server *asynq.Server
	store  db.Store
	mailer mail.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor {
	logger := NewLogger()
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCrtical: 10,
				QueueDefault: 5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: logger,
		},
	)
	return &RedisTaskPrcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (processor *RedisTaskPrcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendEmail, processor.ProcessTaskSendEmail)
	mux.HandleFunc(TaskSendPasswordResetTokenEmail, processor.ProcessTaskSendPasswordResetTokenEmail)
	return processor.server.Start(mux)
}
func (processor *RedisTaskPrcessor) Shutdown() {
	processor.server.Shutdown()
}
