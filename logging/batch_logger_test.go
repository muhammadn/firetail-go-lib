package logging

import (
	"encoding/json"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetupLogger(batchChannel chan *[][]byte, maxBatchSize int, maxLogAge time.Duration) *batchLogger {
	batchLogger := NewBatchLogger(maxBatchSize, maxLogAge, "")

	// Replace the batchHandler with a custom one to throw the batches into a queue that we can receive from for testing
	batchLogger.batchHandler = func(b [][]byte) error {
		batchChannel <- &b
		return nil
	}

	return batchLogger
}

func TestOldLogIsSentImmediately(t *testing.T) {
	const MaxLogAge = time.Minute

	batchChannel := make(chan *[][]byte, 2)
	batchLogger := SetupLogger(batchChannel, 1024*512, MaxLogAge)

	// Create a test log entry & enqueue it
	testLogEntry := LogEntry{
		DateCreated: time.Now().UnixMilli() - MaxLogAge.Milliseconds()*2,
	}
	batchLogger.Enqueue(&testLogEntry)

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

func TestBatchesDoNotExceedMaxSize(t *testing.T) {
	const TestLogEntryCount = 1000
	const MaxBatchSize = 1024 * 512 // 512KB in Bytes

	// Buffer our batchChannel with TestLogEntryCount spaces (worst case, each entry ends up in its own batch)
	batchChannel := make(chan *[][]byte, TestLogEntryCount)
	batchLogger := SetupLogger(batchChannel, MaxBatchSize, time.Second)

	// Create a bunch of test entries
	testLogEntries := []*LogEntry{}
	for i := 0; i < TestLogEntryCount; i++ {
		testLogEntries = append(
			testLogEntries,
			&LogEntry{
				DateCreated:   time.Now().UnixMilli(),
				ExecutionTime: rand.Float64() * 100,
				Request: Request{
					Body: "{\"description\":\"This is a test request body\"}",
					Headers: map[string][]string{
						"Content-Type": {"application/json"},
					},
					HTTPProtocol: HTTP2,
					IP:           "8.8.8.8",
					Method:       "POST",
					URI:          "This isn't a real URI",
					Resource:     "This isn't a real resource",
				},
				Response: Response{
					// Create response bodies of varying size so the batch sizes aren't all the same
					Body: strings.Repeat("a", rand.Intn(10000)),
					Headers: map[string][]string{
						"Content-Type": {"text/plain"},
					},
					StatusCode: 200,
				},
			},
		)
	}

	// Enqueue them all
	for _, logEntry := range testLogEntries {
		batchLogger.Enqueue(logEntry)
	}

	// Keep reading until we've seen all the log entries
	logEntriesReceived := 0
	for logEntriesReceived < TestLogEntryCount {
		batch := <-batchChannel

		logEntriesReceived += len(*batch)

		batchSize := 0
		for _, logBytes := range *batch {
			batchSize += len(logBytes)
		}

		// No batch should ever be bigger than the MaxBatchSize
		assert.GreaterOrEqual(t, MaxBatchSize, batchSize)
	}

	// We should receive exactly the same number of log entries as we put in
	assert.Equal(t, TestLogEntryCount, logEntriesReceived)

	// There should also be no batches left
	assert.Equal(t, 0, len(batchChannel))
}

func TestOldLogTriggersBatch(t *testing.T) {
	const ExpectedLogEntryCount = 10
	const MaxLogAge = time.Minute

	batchChannel := make(chan *[][]byte, 2)
	batchLogger := SetupLogger(batchChannel, 1024*512, MaxLogAge)

	// Create ExpectedLogEntryCount-1 test log entries (the last one will trigger a batch)
	testLogEntries := []*LogEntry{}
	for i := 0; i < ExpectedLogEntryCount-1; i++ {
		testLogEntries = append(
			testLogEntries,
			&LogEntry{
				DateCreated: time.Now().UnixMilli(),
			},
		)
	}

	// Enqueue the first group of test entries (all younger than MaxLogAge)
	for _, logEntry := range testLogEntries {
		batchLogger.Enqueue(logEntry)
	}

	// Assert that there's no batches ready yet
	assert.Equal(t, 0, len(batchChannel))

	// Create a test log entry that's older than MaxLogAge & enqueue it
	oldLogEntry := &LogEntry{
		DateCreated: time.Now().UnixMilli() - MaxLogAge.Milliseconds()*2,
	}
	testLogEntries = append(testLogEntries, oldLogEntry)
	batchLogger.Enqueue(oldLogEntry)

	// There should then be a batch in the channel for us to receive
	batch := <-batchChannel

	// The batch channel should now be empty, as it should only have had one batch in it
	assert.Equal(t, 0, len(batchChannel))

	// Assert that the batch had the correct number of log entries in it
	require.Equal(t, ExpectedLogEntryCount, len(*batch))

	// Create the expected batch from our test log entries
	expectedBatch := [][]byte{}
	for _, logEntry := range testLogEntries {
		logEntryBytes, err := json.Marshal(logEntry)
		require.Nil(t, err)
		expectedBatch = append(expectedBatch, logEntryBytes)
	}

	// Assert that the batch has all the same byte slices as the expected batch
	require.ElementsMatch(t, expectedBatch, *batch)
}
