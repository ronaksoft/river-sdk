package domain

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// QueueUpdates update envelop queue for network debouncer
type QueueUpdates struct {
	mx     sync.Mutex
	items  []*msg.UpdateContainer
	length int
}

// NewQueueUpdates create new instance
func NewQueueUpdates() *QueueUpdates {
	return &QueueUpdates{
		items: make([]*msg.UpdateContainer, 0),
	}
}

// PushMany insert items to queue
func (q *QueueUpdates) PushMany(m []*msg.UpdateContainer) {
	q.mx.Lock()
	q.items = append(q.items, m...)
	q.length += len(m)
	q.mx.Unlock()
}

// Push insert item to queue
func (q *QueueUpdates) Push(m *msg.UpdateContainer) {
	q.mx.Lock()
	q.items = append(q.items, m)
	q.length++
	q.mx.Unlock()
}

// Pop pickup item from queue
func (q *QueueUpdates) Pop() (*msg.UpdateContainer, error) {
	if q.length > 0 {
		q.mx.Lock()
		m := q.items[0]
		q.length--
		q.items = q.items[1:]
		q.mx.Unlock()
		return m, nil
	}
	return nil, ErrDoesNotExists
}

// PopAll pick all items from queue
func (q *QueueUpdates) PopAll() []*msg.UpdateContainer {
	q.mx.Lock()
	m := q.items[:]
	q.length = 0
	q.items = make([]*msg.UpdateContainer, 0)
	q.mx.Unlock()
	return m
}

// Clear queue
func (q *QueueUpdates) Clear() {
	q.mx.Lock()
	q.length = 0
	q.items = make([]*msg.UpdateContainer, 0)
	q.mx.Unlock()
}

// Length  size of queue
func (q *QueueUpdates) Length() int {
	return q.length
}

// GetRawItems return underlying slice
func (q *QueueUpdates) GetRawItems() []*msg.UpdateContainer {
	return q.items
}
