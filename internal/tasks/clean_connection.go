package tasks

import (
	"context"
	"demo/internal/server/websocket"
	"github.com/hibiken/asynq"
)

const CleanConnection = "websocket:clean_connection"

func NewCleanConnection() *asynq.Task {
	return asynq.NewTask(CleanConnection, make([]byte, 1))
}

func HandlerCleanConnectionTask(ctx context.Context, t *asynq.Task) error {
	websocket.ClearTimeoutConnections()
	return nil
}
