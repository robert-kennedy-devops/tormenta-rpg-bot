package bot

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UpdateHandlers struct {
	OnMessage  func(*tgbotapi.Message)
	OnCallback func(*tgbotapi.CallbackQuery)
}

// StartUpdateWorkerPool consumes Telegram updates and dispatches them to a
// configurable worker pool. It preserves compatibility with current handlers.
func StartUpdateWorkerPool(updates tgbotapi.UpdatesChannel, workers int, h UpdateHandlers) {
	if workers < 1 {
		workers = 1
	}

	jobs := make(chan tgbotapi.Update, workers*64)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for up := range jobs {
				// Panic isolation to avoid killing workers in production.
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("[UpdateWorker %d] panic recovered: %v", workerID, r)
						}
					}()

					if up.Message != nil && h.OnMessage != nil {
						h.OnMessage(up.Message)
					}
					if up.CallbackQuery != nil && h.OnCallback != nil {
						h.OnCallback(up.CallbackQuery)
					}
				}()
			}
		}(i + 1)
	}

	go func() {
		for up := range updates {
			jobs <- up
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		log.Printf("[UpdateWorker] pool stopped")
	}()
}
