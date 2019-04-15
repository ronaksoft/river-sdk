package network

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"sync"
)

/*
   Creation Time: 2019 - Apr - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

// MessageQueue message envelop queue for network debouncer
type UpdateQueue struct {
	mx    sync.Mutex
	items chan *msg.UpdateContainer
}

// NewQueueMessages create new instance
func NewUpdateQueue(size int) *UpdateQueue {
	return &UpdateQueue{
		items: make(chan *msg.UpdateContainer, size),
	}
}

// PushMany insert items to queue
func (q *UpdateQueue) PushMany(m []*msg.UpdateContainer) {
	q.mx.Lock()
	for idx := range m {
		q.items <- m[idx]
	}
	q.mx.Unlock()
}

// Push insert item to queue
func (q *UpdateQueue) Push(m *msg.UpdateContainer) {
	q.mx.Lock()
	q.items <- m
	q.mx.Unlock()
}

// Pop pickup item from queue
func (q *UpdateQueue) Pop() (*msg.UpdateContainer, error) {
	q.mx.Lock()
	defer q.mx.Unlock()
	select {
	case m := <-q.items:
		return m, nil
	default:
	}
	return nil, domain.ErrDoesNotExists
}

// PopAll pick all items from queue
func (q *UpdateQueue) PopAll() []*msg.UpdateContainer {
	q.mx.Lock()
	defer q.mx.Unlock()

	itemsCount := len(q.items)
	items := make([]*msg.UpdateContainer, 0, itemsCount)
	for i := 0; i < itemsCount; i++ {
		items = append(items, <-q.items)
	}
	return items
}

// Length  size of queue
func (q *UpdateQueue) Length() int {
	q.mx.Lock()
	itemsCount := len(q.items)
	q.mx.Unlock()
	return itemsCount
}
