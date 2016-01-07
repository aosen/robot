package scheduler

import (
	"fmt"
	"testing"

	"github.com/aosen/robot"
)

func TestQueueScheduler(t *testing.T) {
	var r *robot.Request
	r = robot.NewRequest("http://baidu.com", "html", "", "GET", "", nil, nil, nil, nil)
	fmt.Printf("%v\n", r)

	var s *QueueScheduler
	s = NewQueueScheduler(false)

	s.Push(r)
	var count int = s.Count()
	if count != 1 {
		t.Error("count error")
	}
	fmt.Println(count)

	var r1 *robot.Request
	r1 = s.Poll()
	if r1 == nil {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)

	// remove duplicate
	s = NewQueueScheduler(true)

	r2 := robot.NewRequest("http://qq.com", "html", "", "GET", "", nil, nil, nil, nil)
	s.Push(r)
	s.Push(r2)
	s.Push(r)
	count = s.Count()
	if count != 2 {
		t.Error("count error")
	}
	fmt.Println(count)

	r1 = s.Poll()
	if r1 == nil {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)
	r1 = s.Poll()
	if r1 == nil {
		t.Error("poll error")
	}
	fmt.Printf("%v\n", r1)
}
