package logging

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var batchLogger *BatchLogger

func SetupLogger(batchChannel chan *[][]byte, maxBatchSize int, maxLogAge time.Duration) {
	batchLogger = NewBatchLogger(maxBatchSize, maxLogAge, "")

	// Replace the batchHandler with a custom one to throw the batches into a queue that we can receive from for testing
	batchLogger.batchHandler = func(b [][]byte) error {
		log.Printf("Queueing batch of size %d\n", len(b))
		batchChannel <- &b
		return nil
	}
}

func TestOldLogIsSentImmediately(t *testing.T) {
	batchChannel := make(chan *[][]byte, 2)
	SetupLogger(batchChannel, 1024^3, time.Minute)

	// Create a test log entry & enqueue it
	testLogEntry := LogEntry{
		DateCreated: 0,
	}
	go batchLogger.Enqueue(&testLogEntry)

	// There should then be a batch in the channel for us to receive
	batch := <-batchChannel

	// Channel should be empty now, as it should only have had one batch in it
	assert.Equal(t, 0, len(batchChannel))

	// Marshal the testLogEntry to get our expected bytes
	expectedLogEntryBytes, err := json.Marshal(testLogEntry)
	require.Nil(t, err)

	// Assert the batch had one log entry in it, which matches our test log entry's bytes
	require.Equal(t, 1, len(*batch))
	assert.Equal(t, expectedLogEntryBytes, (*batch)[0])
}
