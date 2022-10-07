package logging

import (
	"encoding/json"
	"log"
	"time"
)

// WIP
type batchLogger struct {
	queue        chan *LogEntry       // A channel down which LogEntrys will be queued to be sent to Firetail
	maxBatchSize int                  // The maximum size of a batch in bytes
	maxLogAge    time.Duration        // The maximum age of a log item to hold onto
	batchHandler func([][]byte) error // A handler that takes a batch of log entries as a slice of slices of bytes & sends them to Firetail
}

func NewBatchLogger(maxBatchSize int, maxLogAge time.Duration, loggingEndpoint string) *batchLogger {
	newLogger := &batchLogger{
		queue:        make(chan *LogEntry),
		maxBatchSize: maxBatchSize,
		maxLogAge:    maxLogAge,
	}

	newLogger.batchHandler = func(batchBytes [][]byte) error {
		// TODO: send to Firetail. If there's an err, we should re-queue
		log.Printf("Sending %d log(s) to '%s'...", len(batchBytes), loggingEndpoint)
		for i, entry := range batchBytes {
			log.Printf("Entry #%d: %s\n", i, string(entry))
		}
		return nil
	}

	go newLogger.worker()

	return newLogger
}

// Enqueue enqueues a logentry to be batched & sent to Firetail. Should normally be run in a new goroutine as it blocks until another routine receives from l.queue.
func (l *batchLogger) Enqueue(logEntry *LogEntry) {
	l.queue <- logEntry
}

func (l *batchLogger) worker() {
	currentBatch := [][]byte{}
	currentBatchSize := 0
	var oldestEntryCreatedAt *time.Time

	for {
		batchIsReady := false

		// Read a new entry from the queue if there's one available
		select {
		case newEntry := <-l.queue:
			// Marshal the entry to bytes...
			entryBytes, err := json.Marshal(newEntry)
			if err != nil {
				panic(err)
			}

			// If it's too big to add to the batch, re-enqueue it, flag the batch is ready to send & break from this case
			if len(entryBytes)+currentBatchSize > l.maxBatchSize {
				go l.Enqueue(newEntry)
				batchIsReady = true
				break
			}

			// Append it to the batch & increment the currentBatchSize appropriately
			currentBatch = append(currentBatch, entryBytes)
			currentBatchSize += len(entryBytes)

			// If the new entry is older than the oldest currently in the batch, we update oldestEntryCreatedAt
			if oldestEntryCreatedAt == nil || newEntry.DateCreated < oldestEntryCreatedAt.UnixMilli() {
				createdAt := time.UnixMilli(newEntry.DateCreated)
				oldestEntryCreatedAt = &createdAt
			}
		default:
			// If there's no new entry available, just break
			break
		}

		// If the oldest entry in the currentBatch was logged long enough ago, then the currentBatch is ready to send
		if oldestEntryCreatedAt != nil && time.Since(*oldestEntryCreatedAt) > l.maxLogAge {
			batchIsReady = true
		}

		if batchIsReady {
			go l.batchHandler(currentBatch)
			currentBatch = [][]byte{}
			currentBatchSize = 0
			oldestEntryCreatedAt = nil
		}
	}
}
