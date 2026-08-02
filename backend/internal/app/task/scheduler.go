package task

import (
	"context"
	"log"
	"time"
)

func (service *Service) RunNotificationLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if _, err := service.NotifyDueTasks(ctx, now); err != nil {
				log.Printf("notify due tasks failed: %v", err)
			}
		}
	}
}
