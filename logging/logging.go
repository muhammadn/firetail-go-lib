package logging

import (
	"sync/atomic"
	"time"
)

// WIP
type Logger struct {
	queue        chan *LogEntry // A channel down which LogEntrys will be queued to be sent
	queueSize    *int64         // The number of log entries in the queue
	endpoint     string         // The endpoint to which logs will be sent in batches
	maxBatchSize int            // The maximum size of a batch
	maxLogAge    time.Duration  // The maximum age of a log item to hold onto
}

func (l *Logger) EnqueueLogEntry(logEntry *LogEntry) {
	l.queue <- logEntry
	atomic.AddInt64(l.queueSize, 1)
}
