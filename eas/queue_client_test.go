package eas

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pai-eas/eas-golang-sdk/eas/types"
)

const (
	QueueEndpoint  = "http://localhost:3030"
	InputQueueName = ""
	SinkQueueName  = "sink"
	QueueToken     = ""
)

type QueueClientTestCase struct {
	inputQueue *QueueClient
	sinkQueue  *QueueClient
}

func assertEqual(t *testing.T, a, b interface{}) {
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func getQueueClient(t *testing.T, opts ...QueueOption) *QueueClientTestCase {
	testCase := &QueueClientTestCase{}
	var err error
	opts = append(opts, WithBasePath(""))
	testCase.inputQueue, err = NewQueueClient(QueueEndpoint, InputQueueName, QueueToken, opts...)
	assertNoError(t, err)
	testCase.sinkQueue, err = NewQueueClient(QueueEndpoint, SinkQueueName, QueueToken, opts...)
	assertNoError(t, err)
	return testCase
}

func getRearQueueClient(t *testing.T, opts ...QueueOption) *QueueClientTestCase {
	testCase := &QueueClientTestCase{}
	var err error
	opts = append(opts, WithBasePath(""), WithExtraHeaders(map[string]string{HeaderAccessRear: "true"}))
	testCase.inputQueue, err = NewQueueClient(QueueEndpoint, InputQueueName, QueueToken, opts...)
	assertNoError(t, err)
	testCase.sinkQueue, err = NewQueueClient(QueueEndpoint, SinkQueueName, QueueToken, opts...)
	assertNoError(t, err)
	return testCase
}

func (c *QueueClientTestCase) truncate(t *testing.T) {
	attrs, err := c.inputQueue.Attributes()
	assertNoError(t, err)
	if index, ok := attrs["stream.lastEntry"]; ok {
		idx, _ := strconv.ParseUint(index, 10, 64)
		c.inputQueue.Truncate(context.Background(), idx+1)
	}

	attrs, err = c.sinkQueue.Attributes()
	assertNoError(t, err)
	if index, ok := attrs["stream.lastEntry"]; ok {
		idx, _ := strconv.ParseUint(index, 10, 64)
		c.sinkQueue.Truncate(context.Background(), idx+1)
	}

}

func TestTruncate(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	latestIndex := uint64(0)
	for i := 0; i < 10; i++ {
		index, _, err := c.sinkQueue.Put(context.Background(), []byte("abc"), types.Tags{})
		assertNoError(t, err)
		latestIndex = index
	}

	c.sinkQueue.Truncate(context.Background(), latestIndex+1)

	attrs, err := c.sinkQueue.Attributes()
	assertNoError(t, err)

	assertEqual(t, attrs["stream.length"], "0")
}

func TestQueueGetByRequestId(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	_, requestId, err := c.sinkQueue.Put(context.Background(), []byte("abc"), types.Tags{})
	assertNoError(t, err)

	list, err := c.sinkQueue.GetByRequestId(context.Background(), requestId)
	assertNoError(t, err)

	assertEqual(t, len(list), 1)
	assertEqual(t, string(list[0].Data), "abc")
}

func TestQueueGetByIndex(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	index, _, err := c.sinkQueue.Put(context.Background(), []byte("abc"), types.Tags{})
	assertNoError(t, err)

	list, err := c.sinkQueue.GetByIndex(context.Background(), index)
	assertNoError(t, err)

	assertEqual(t, len(list), 1)
	assertEqual(t, string(list[0].Data), "abc")
}

func TestRearQueueGetByRequestId(t *testing.T) {
	c := getRearQueueClient(t)

	c.truncate(t)

	_, requestId, err := c.sinkQueue.Put(context.Background(), []byte("abc"), types.Tags{})
	assertNoError(t, err)

	list, err := c.sinkQueue.GetByRequestId(context.Background(), requestId)
	assertNoError(t, err)

	assertEqual(t, len(list), 1)
	assertEqual(t, string(list[0].Data), "abc")
}

func TestWatchWithAutoCommit(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	for i := 0; i < 10; i++ {
		_, _, err := c.sinkQueue.Put(context.Background(), []byte(strconv.Itoa(i)), types.Tags{})
		assertNoError(t, err)
	}

	watcher, err := c.sinkQueue.Watch(context.Background(), 0, 5, false, true)
	assertNoError(t, err)

	for i := 0; i < 10; i++ {
		df := <-watcher.FrameChan()
		assertEqual(t, string(df.Data), strconv.Itoa(i))
	}

	watcher.Close()

	time.Sleep(2 * time.Second)

	attrs, err := c.sinkQueue.Attributes()
	assertNoError(t, err)
	assertEqual(t, attrs["stream.length"], "0")
}

func TestWatchWithManualCommit(t *testing.T) {
	c := getQueueClient(t)

	c.truncate(t)

	for i := 0; i < 10; i++ {
		_, _, err := c.sinkQueue.Put(context.Background(), []byte(strconv.Itoa(i)), types.Tags{})
		assertNoError(t, err)
	}

	watcher, err := c.sinkQueue.Watch(context.Background(), 0, 5, false, false)
	assertNoError(t, err)

	for i := 0; i < 10; i++ {
		df := <-watcher.FrameChan()
		err := c.sinkQueue.Commit(context.Background(), df.Index.Uint64())
		assertNoError(t, err)
		assertEqual(t, string(df.Data), strconv.Itoa(i))
	}

	watcher.Close()

	time.Sleep(2 * time.Second)

	attrs, err := c.sinkQueue.Attributes()
	assertNoError(t, err)
	assertEqual(t, attrs["stream.length"], "0")
}

func TestWatchWithReconnect(t *testing.T) {
	total := int32(0)
	c := getQueueClient(t, withBreakGenerator(
		func() bool {
			return atomic.LoadInt32(&total) == 60
		},
	))

	c.truncate(t)

	watcher, err := c.sinkQueue.Watch(context.Background(), 0, 5, false, false)
	assertNoError(t, err)

	go func() {
		for i := 0; i < 100; i++ {
			_, _, err := c.sinkQueue.Put(context.Background(), []byte(strconv.Itoa(i)), types.Tags{})
			assertNoError(t, err)
			atomic.AddInt32(&total, 1)
			time.Sleep(time.Microsecond * 10)
		}
		time.Sleep(time.Second * 2)
		watcher.Close()
	}()

	ch := watcher.FrameChan()
	for df := range ch {
		err := c.sinkQueue.Commit(context.Background(), df.Index.Uint64())
		if err != nil {
			fmt.Printf("commit id: %v failed: %v", df.Index, err)
		}
		assertNoError(t, err)
	}

	time.Sleep(time.Second * 2)
	attrs, err := c.sinkQueue.Attributes()
	assertNoError(t, err)
	assertEqual(t, attrs["stream.length"], "0")
}

func TestDemoWatch(t *testing.T) {
	const (
		QueueEndpoint   = ""
		inputQueueName  = ""
		outputQueueName = ""
		QueueToken      = ""
	)
	inuptQueue, err := NewQueueClient(QueueEndpoint, inputQueueName, QueueToken)

	// truncate all messages in the queue
	attrs, err := inuptQueue.Attributes()
	if index, ok := attrs["stream.lastEntry"]; ok {
		idx, _ := strconv.ParseUint(index, 10, 64)
		inuptQueue.Truncate(context.Background(), idx+1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// create a goroutine to send messages to the queue
	go func() {
		i := 0
		for {
			select {
			case <-time.NewTicker(time.Microsecond * 1).C:
				_, _, err := inuptQueue.Put(context.Background(), []byte(strconv.Itoa(i)), types.Tags{})
				if err != nil {
					fmt.Printf("Error occured, retry to handle it: %v\n", err)
				}
				i += 1
			case <-ctx.Done():
				break
			}
		}
	}()

	outputQueue, err := NewQueueClient(QueueEndpoint, outputQueueName, QueueToken)
	// create a watcher to watch the messages from the queue
	watcher, err := outputQueue.Watch(context.Background(), 0, 5, false, false)
	if err != nil {
		fmt.Printf("Failed to create a watcher to watch the queue: %v\n", err)
		return
	}
	// read messages from the queue and commit manually
	for i := 0; i < 100; i++ {
		df := <-watcher.FrameChan()
		fmt.Printf("Get data: %v\n", string(df.Data))
		err := outputQueue.Commit(context.Background(), df.Index.Uint64())
		if err != nil {
			fmt.Printf("Failed to commit index: %v(%v)\n", df.Index, err)
		}
	}

	// everything is done, close the watcher
	watcher.Close()
	cancel()
}
