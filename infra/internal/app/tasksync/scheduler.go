package tasksync

import (
	"context"
	"log"
	"time"
)

func (service *Service) RunSyncLoop(ctx context.Context, interval time.Duration) {
	service.syncAndLog(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			service.syncAndLog(ctx)
		}
	}
}

func (service *Service) syncAndLog(ctx context.Context) {
	tasks, err := service.SyncTasks(ctx)
	if err != nil {
		log.Printf("notion task sync failed: %v", err)
		return
	}

	log.Printf("notion task sync completed: tasks=%d", len(tasks))
}
