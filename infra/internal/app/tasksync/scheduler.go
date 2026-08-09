package tasksync

import (
	"context"
	"log"
	"time"

	"github.com/med-000/med-control/shared/timeutil"
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

func (service *Service) RunScheduleCompletionLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	var lastCompletedDate string
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			now = now.In(timeutil.JST())
			if now.Hour() != 1 {
				continue
			}
			completedDate := now.Format(time.DateOnly)
			if completedDate == lastCompletedDate {
				continue
			}

			tasks, err := service.CompleteScheduleItems(ctx, now)
			if err != nil {
				log.Printf("schedule item completion failed: %v", err)
				continue
			}
			lastCompletedDate = completedDate
			log.Printf("schedule item completion completed: tasks=%d", len(tasks))
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
