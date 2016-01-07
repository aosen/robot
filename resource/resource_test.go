package resource

import "testing"

func TestSimpleManage(t *testing.T) {
	rm := NewSimpleManage(1)
	rm.GetOne()
	println("incr")
	rm.FreeOne()
	println("decr")
	rm.GetOne()
	println("incr")
}
