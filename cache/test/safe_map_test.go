package test

import (
	"github.com/dongma/imola/cache/concurrency/sync"
	"testing"
	"time"
)

func TestSafeMap_LoadOrStore(t *testing.T) {
	safeMap := &sync.SafeMap[string, string]{
		Data: make(map[string]string),
	}

	// test-case1, 有多个goroutine读写问题，两个都是false
	/* === RUN  TestSafeMap_LoadOrStore
		safe_map_test.go:21: goroutine2 value:  value2 false
		safe_map_test.go:16: goroutine1 value:  value1 false
	--- PASS: TestSafeMap_LoadOrStore (4.58s)
	PASS */

	// test-case2, double-check问题，修复后, goroutine的结果符合预期
	/* === RUN   TestSafeMap_LoadOrStore
	    safe_map_test.go:28: goroutine2 value:  value2 false
	    safe_map_test.go:23: goroutine1 value:  value2 true
	--- PASS: TestSafeMap_LoadOrStore (6.95s)
	PASS
	*/

	go func() {
		val, ok := safeMap.LoadOrStore("key1", "value1")
		t.Log("goroutine1 value: ", val, ok)
	}()

	go func() {
		val, ok := safeMap.LoadOrStore("key1", "value2")
		t.Log("goroutine2 value: ", val, ok)
	}()

	time.Sleep(time.Second)
}
