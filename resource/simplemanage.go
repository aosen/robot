/*
Author: Aosen
Data: 2016-01-07
QQ: 316052486
Desc: 通过chanel来实现的简易资源管理器
*/
package resource

// ResourceManageChan inherits the ResourceManage interface.
// In spider, ResourceManageChan manage resource of Coroutine to crawl page.
type SimpleManage struct {
	capnum uint
	mc     chan uint
}

// NewResourceManageChan returns initialized ResourceManageChan object which contains a resource pool.
// The num is the resource limit.
func NewSimpleManage(num uint) *SimpleManage {
	mc := make(chan uint, num)
	return &SimpleManage{mc: mc, capnum: num}
}

// The GetOne apply for one resource.
// If resource pool is empty, current coroutine will be blocked.
func (self *SimpleManage) GetOne() {
	self.mc <- 1
}

// The FreeOne free resource and return it to resource pool.
func (self *SimpleManage) FreeOne() {
	<-self.mc
}

// The Has query for how many resource has been used.
func (self *SimpleManage) Has() uint {
	return uint(len(self.mc))
}

// The Left query for how many resource left in the pool.
func (self *SimpleManage) Left() uint {
	return self.capnum - uint(len(self.mc))
}
